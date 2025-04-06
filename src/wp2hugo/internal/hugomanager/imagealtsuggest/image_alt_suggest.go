package imagealtsuggest

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/adrg/frontmatter"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/llmhelper"
	"github.com/openai/openai-go"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"os"
	"path"
	"regexp"
	"strings"
)

const _imageAltSystemPrompt = `
Provide a functional, objective description of the image in under 30 words.
Focus on the main object, its action, and context. Transcribe important text if present,
avoiding quotation marks. Do not start with "The image.`

// Parse "{{< figure align=aligncenter width=768 src="Cedar_trail_waterfall-768x1024.jpg" alt="" >}}"
// and extract "src" and "alt" attributes using regular expressions
var _figureShortCodeRegEx = regexp.MustCompile(`{{<\s*?figure.*?>\s*?}}`)
var _figureShortCodeSrcRegEx = regexp.MustCompile(`src=['"](.*?)['"]`)
var _figureShortCodeAltRegEx = regexp.MustCompile(`alt=['"](.*?)['"]`)

type Result struct {
	numImageWithAlt    int
	numImageMissingAlt int
	numImageUpdated    int
}

func (r Result) NumImageWithAlt() int {
	return r.numImageWithAlt
}

func (r Result) NumImageMissingAlt() int {
	return r.numImageMissingAlt
}

func (r Result) NumImageUpdated() int {
	return r.numImageUpdated
}

func ProcessFile(ctx context.Context, mdFilePath string, updateInline bool) (*Result, error) {
	f, err := os.OpenFile(mdFilePath, os.O_RDONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", mdFilePath, err)
	}

	defer func() {
		_ = f.Close()
	}()
	var frontmatterData any
	log.Debug().
		Str("mdFilePath", mdFilePath).
		Msg("Parsing frontmatter")
	mdBodyBytes, err := frontmatter.Parse(f, &frontmatterData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter in file %s: %w", mdFilePath, err)
	}
	if len(mdBodyBytes) == 0 {
		log.Debug().
			Str("mdFilePath", mdFilePath).
			Msg("No markdown body found, skipping")
		return &Result{}, nil
	}

	mdBody := string(mdBodyBytes)
	numImageWithAlt := 0
	numImageMissingAlt := 0
	numUpdated := 0
	// Check if the body contains images
	// and if they are missing alt text
	figureMatches := _figureShortCodeRegEx.FindAllString(mdBody, -1)
	for _, figureMatch := range figureMatches {
		srcMatches := _figureShortCodeSrcRegEx.FindAllStringSubmatch(figureMatch, -1)
		if len(srcMatches) == 0 {
			log.Warn().
				Str("mdFilePath", mdFilePath).
				Msg("No src attribute found in figure shortcode")
			continue
		}
		src := srcMatches[0][1]
		if len(src) == 0 {
			log.Warn().
				Str("mdFilePath", mdFilePath).
				Msg("Empty src attribute found in figure shortcode")
			continue
		}

		altMatches := _figureShortCodeAltRegEx.FindAllStringSubmatch(figureMatch, -1)
		if len(altMatches) == 0 || len(altMatches[0]) < 2 || string(altMatches[0][1]) == "" {
			log.Warn().
				Str("mdFilePath", mdFilePath).
				Msg("No alt attribute found in figure shortcode")
			numImageMissingAlt++
			if updateInline {
				// This assumes that the images are in the same directory as the markdown file
				// Which is not historically true.
				// The images could be in the static directory or could be in another
				// directory being referenced via URL of the post that owns that image.
				// For now, this does not handle those two cases.
				imgFilePath := path.Join(path.Dir(mdFilePath), src)

				// Get 100 characters before and after the "figureMatch"
				// to get the context of the image
				start := strings.Index(mdBody, figureMatch)
				start = max(0, start-100)
				end := strings.Index(mdBody, figureMatch) + len(figureMatch) + 100
				end = min(len(mdBody), end)
				contextForImage := mdBody[start:end]
				alt, err := getAlt(ctx, imgFilePath, contextForImage)
				if err != nil {
					return nil, fmt.Errorf("failed to get alt text for image %s: %w", src, err)
				}
				figureMatchWithAlt := _figureShortCodeAltRegEx.ReplaceAllString(figureMatch, fmt.Sprintf(`alt="%s"`, *alt))
				if figureMatchWithAlt == figureMatch { // figureMatch is missing "alt" attribute altogether
					figureMatchWithAlt = strings.Replace(figureMatch, ">", fmt.Sprintf(` alt="%s">`, *alt), 1)
				}
				if err := replaceInFile(mdFilePath, figureMatch, figureMatchWithAlt); err != nil {
					return nil, fmt.Errorf("failed to update file %s: %w", mdFilePath, err)
				}
				numUpdated++
			}
		} else {
			numImageWithAlt++
		}
	}

	return &Result{
		numImageWithAlt:    numImageWithAlt,
		numImageMissingAlt: numImageMissingAlt,
		numImageUpdated:    numUpdated,
	}, nil
}

func replaceInFile(filePath string, old string, new string) error {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	modifiedBody := strings.Replace(string(fileData), old, new, 1)
	if err := os.WriteFile(filePath, []byte(modifiedBody), 0o644); err != nil {
		return fmt.Errorf("failed to write updated file %s: %w", filePath, err)
	}

	return nil
}

func getAlt(ctx context.Context, imgPath string, textAroundImage string) (*string, error) {
	data, err := os.ReadFile(imgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file %s: %w", imgPath, err)
	}

	// Convert to base64
	base64Data := base64.URLEncoding.EncodeToString(data)
	var altText *string

	if len(base64Data) > 100_000 && strings.TrimSpace(textAroundImage) != "" {
		log.Warn().
			Str("imgPath", imgPath).
			Msg("Generating alt using the text around the image instead")
		altText, err = llmhelper.CallLLM(ctx, openai.ChatModelGPT4o, _imageAltSystemPrompt,
			"text around image: "+textAroundImage)
	} else {
		altText, err = llmhelper.CallLLM(ctx, openai.ChatModelGPT4o, _imageAltSystemPrompt,
			"text around image: "+textAroundImage+"\n image: "+base64Data)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get alt text for image %s: %w", imgPath, err)
	}

	altText = lo.ToPtr(strings.ReplaceAll(*altText, `"`, `'`))
	altText = lo.ToPtr(strings.ReplaceAll(*altText, "  ", " "))
	log.Debug().
		Str("imgPath", imgPath).
		Str("altText", *altText).
		Msg("Generated alt text for image")
	return altText, nil
}
