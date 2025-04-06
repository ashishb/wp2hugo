package imagealtsuggest

import (
	"context"
	"fmt"
	"github.com/adrg/frontmatter"
	"github.com/rs/zerolog/log"
	"os"
	"regexp"
)

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

func ProcessFile(ctx context.Context, path string, updateInline bool) (*Result, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}

	defer func() {
		_ = f.Close()
	}()
	var frontmatterData any
	log.Debug().
		Str("path", path).
		Msg("Parsing frontmatter")
	mdBody, err := frontmatter.Parse(f, &frontmatterData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter in file %s: %w", path, err)
	}
	if len(mdBody) == 0 {
		log.Debug().
			Str("path", path).
			Msg("No markdown body found, skipping")
		return &Result{}, nil
	}

	numImageWithAlt := 0
	numImageMissingAlt := 0
	// Check if the body contains images
	// and if they are missing alt text
	figureMatches := _figureShortCodeRegEx.FindAll(mdBody, -1)
	for _, figureMatch := range figureMatches {
		srcMatches := _figureShortCodeSrcRegEx.FindAllSubmatch(figureMatch, -1)
		if len(srcMatches) == 0 {
			log.Warn().
				Str("path", path).
				Msg("No src attribute found in figure shortcode")
			continue
		}
		src := string(srcMatches[0][1])
		if len(src) == 0 {
			log.Warn().
				Str("path", path).
				Msg("Empty src attribute found in figure shortcode")
			continue
		}

		altMatches := _figureShortCodeAltRegEx.FindAllSubmatch(figureMatch, -1)
		if len(altMatches) == 0 || len(altMatches[0]) < 2 || string(altMatches[0][1]) == "" {
			log.Warn().
				Str("path", path).
				Msg("No alt attribute found in figure shortcode")
			numImageMissingAlt++
		} else {
			numImageWithAlt++
		}
	}

	if updateInline {
		return nil, fmt.Errorf("inline update not supported yet")
	}

	return &Result{
		numImageWithAlt:    numImageWithAlt,
		numImageMissingAlt: numImageMissingAlt,
		numImageUpdated:    0,
	}, nil
}
