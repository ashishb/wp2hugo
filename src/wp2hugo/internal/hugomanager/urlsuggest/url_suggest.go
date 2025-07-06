package urlsuggest

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/frontmatterhelper"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/rs/zerolog/log"
)

func ProcessFile(path string, updateInline bool) (*string, error) {
	if !strings.HasSuffix(path, ".md") {
		return nil, errors.New("file is not a markdown file")
	}

	if !utils.FileExists(path) {
		return nil, fmt.Errorf("file does not exist: %s", path)
	}
	matter, err := frontmatterhelper.GetSelectiveFrontMatter(path)
	if err != nil {
		return nil, err
	}

	if matter.URL != "" && matter.URL != "/" {
		log.Debug().
			Str("path", path).
			Msg("front matter URL is present")
		return &matter.URL, nil
	}
	// We are only interested in unpublished posts
	inFuture, err := matter.IsInFuture()
	if err != nil {
		log.Error().
			Err(err).
			Str("path", path).
			Str("date", matter.PublishDate).
			Msg("Error parsing date")
		return nil, err
	}
	unpublished := matter.IsDraft() || inFuture
	if !unpublished {
		log.Debug().
			Str("path", path).
			Msg("post is published")
		return &matter.URL, nil
	}

	url, err := suggestURL(*matter, path)
	if err != nil {
		log.Error().
			Err(err).
			Str("path", path).
			Msg("Error suggesting URL")
		return nil, err
	}
	log.Info().
		Str("path", path).
		Str("url", *url).
		Msg("Suggested URL")

	if updateInline {
		// Get all the fields this time
		// Update the URL field
		// Write the updated front matter back to the file
		if err := frontmatterhelper.UpdateFrontmatter(path, "url", *url); err != nil {
			log.Error().
				Err(err).
				Str("path", path).
				Str("url", *url).
				Msg("Error updating front matter")
			return nil, err
		}
	}
	return url, nil
}

func suggestURL(matter frontmatterhelper.FrontMatter, path string) (*string, error) {
	prefix := getPrefix(matter)
	suffix := getSuffix(matter, path)
	if prefix == "" {
		return nil, errors.New("no prefix found")
	}
	if suffix == "" {
		return nil, errors.New("no suffix found")
	}
	prefix = normalize(prefix)
	suffix = normalize(suffix)
	// Avoid stutter
	suffix = strings.TrimSuffix(suffix, prefix)

	url := fmt.Sprintf("/%s/%s/", normalize(prefix), normalize(suffix))
	return &url, nil
}

func getSuffix(matter frontmatterhelper.FrontMatter, path string) string {
	if matter.Title != "" {
		return matter.Title
	}
	filename := filepath.Base(path)
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

// getPrefix returns the first category or tag that is not generic.
// It might return an empty string if no category or tag is found.
func getPrefix(matter frontmatterhelper.FrontMatter) string {
	genericCategories := map[string]bool{
		"uncategorized": true,
		"all":           true,
	}

	for _, cat := range matter.Categories {
		if !genericCategories[strings.ToLower(cat)] {
			return cat
		}
	}

	for _, tag := range matter.Tags {
		if !genericCategories[strings.ToLower(tag)] {
			return tag
		}
	}
	return ""
}

func normalize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	// Remove all non-alphanumeric characters
	s = regexp.MustCompile("[^a-z0-9-]").ReplaceAllString(s, "")
	// Remove successive hyphens
	s = regexp.MustCompile("-+").ReplaceAllString(s, "-")
	return s
}
