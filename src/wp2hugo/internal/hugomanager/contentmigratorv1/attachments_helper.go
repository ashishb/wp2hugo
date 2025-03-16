package contentmigratorv1

import (
	"fmt"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func hasAttachments(path string) (*bool, error) {
	return hasImageAttachments(path)
}

func hasImageAttachments(path string) (*bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file '%s': %s", path, err)
	}

	if strings.Contains(string(data), "{{< figure") {
		return lo.ToPtr(true), nil
	}

	if strings.Contains(string(data), "![](") {
		return lo.ToPtr(true), nil
	}

	return lo.ToPtr(false), nil
}

func getAllImageAttachmentURLs(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file '%s': %s", path, err)
	}

	// Regex 1 for figure shortcodes
	// {{< figure align=aligncenter width=740 src="image.png" alt="" >}}
	regEx1 := regexp.MustCompile(`{{< figure .* src="(.*?)"`)
	matches := regEx1.FindAllStringSubmatch(string(data), -1)
	images := make([]string, 0)
	for _, match := range matches {
		images = append(images, match[1])
	}

	// Regex 2 for cover images
	//   image: "image.jpg"
	regEx2 := regexp.MustCompile(`image: "(.*?)"`)
	matches = regEx2.FindAllStringSubmatch(string(data), -1)
	for _, match := range matches {
		images = append(images, match[1])
	}

	// Regex 3 for Markdown images
	// ![](/image.png)
	regEx3 := regexp.MustCompile(`!\[.*?]\((.*?)\)`)
	matches = regEx3.FindAllStringSubmatch(string(data), -1)
	for _, match := range matches {
		images = append(images, match[1])
	}

	return lo.Uniq(images), nil
}

func findStaticAttachmentFilePath(markdownFilePath string, attachmentURL string) (*string, error) {
	// Find the static directory
	baseDir := filepath.Dir(markdownFilePath)
	for {
		staticDir := filepath.Join(baseDir, "static")
		if utils.DirExists(staticDir) {
			log.Debug().
				Str("staticDir", staticDir).
				Msg("Found static directory")
			break
		}
		log.Debug().
			Str("staticDir", staticDir).
			Msg("Static directory does not exist, still looking in the parent directory")
		baseDir2 := filepath.Dir(baseDir)
		if baseDir2 == baseDir {
			return nil, fmt.Errorf("static directory not found")
		}
		baseDir = baseDir2
	}

	// Find the attachment file
	attachmentFilePath := filepath.Join(filepath.Join(baseDir, "static"), attachmentURL)
	if utils.FileExists(attachmentFilePath) {
		return &attachmentFilePath, nil
	}

	// Sometimes the attachment URL contains URL encoded characters like %5F for underscore
	attachmentURL, err := url.PathUnescape(attachmentURL)
	if err != nil {
		return nil, fmt.Errorf("error unescaping attachment URL '%s': %s", attachmentURL, err)
	}

	attachmentFilePath = filepath.Join(filepath.Join(baseDir, "static"), attachmentURL)
	if utils.FileExists(attachmentFilePath) {
		return &attachmentFilePath, nil
	}

	return nil, fmt.Errorf("attachment file '%s' not found", attachmentFilePath)
}
