package urlsuggest

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type FrontMatter struct {
	URL         string   `yaml:"url"`
	PublishDate string   `yaml:"date"`
	Draft       string   `yaml:"draft"`
	Title       string   `yaml:"title"`
	Categories  []string `yaml:"category"`
	Tags        []string `yaml:"tag"`
	Summary     string   `yaml:"summary"`
}

func (f *FrontMatter) IsDraft() bool {
	return strings.ToLower(f.Draft) == "true"
}

func (matter *FrontMatter) IsInFuture() (bool, error) {
	if matter.PublishDate == "" {
		return false, nil
	}
	t1, err := time.Parse(time.RFC3339, matter.PublishDate)
	if err != nil {
		return false, err
	}
	return t1.After(time.Now()), nil
}

func ProcessFile(path string, updateInline bool) (*string, error) {
	if !strings.HasSuffix(path, ".md") {
		return nil, fmt.Errorf("file is not a markdown file")
	}

	if !utils.FileExists(path) {
		return nil, fmt.Errorf("file does not exist: %s", path)
	}
	matter, err := GetSelectiveFrontMatter(path)
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
		if err := UpdateFrontmatter(path, "url", *url); err != nil {
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

func UpdateFrontmatter(path string, key string, value string) error {
	fullmatter, restOfTheFile, err := getFullFrontMatter(path)
	if err != nil {
		return err
	}

	fullmatter[key] = value
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	_, err = file.Write([]byte("---\n"))
	if err != nil {
		return err
	}

	yamlEncoder := yaml.NewEncoder(file)
	yamlEncoder.SetIndent(2)
	err = yamlEncoder.Encode(&fullmatter)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte("---\n"))
	if err != nil {
		return err
	}
	_, err = file.Write(restOfTheFile)
	if err != nil {
		return err
	}
	return nil
}

func GetSelectiveFrontMatter(path string) (*FrontMatter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matter FrontMatter
	_, err = frontmatter.Parse(file, &matter)
	if err != nil {
		log.Error().
			Err(err).
			Str("path", path).
			Msg("Error parsing front matter")
		return nil, err
	}
	return &matter, nil
}

func getFullFrontMatter(path string) (map[string]any, []byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var matter map[string]any
	restOfTheFile, err := frontmatter.Parse(file, &matter)
	if err != nil {
		log.Error().
			Err(err).
			Str("path", path).
			Msg("Error parsing front matter")
		return nil, nil, err
	}
	return matter, restOfTheFile, nil
}

func suggestURL(matter FrontMatter, path string) (*string, error) {
	prefix := getPrefix(matter)
	suffix := getSuffix(matter, path)
	if prefix == "" {
		return nil, fmt.Errorf("no prefix found")
	}
	if suffix == "" {
		return nil, fmt.Errorf("no suffix found")
	}
	prefix = normalize(prefix)
	suffix = normalize(suffix)
	// Avoid stutter
	suffix = strings.TrimSuffix(suffix, prefix)

	url := fmt.Sprintf("/%s/%s/", normalize(prefix), normalize(suffix))
	return &url, nil
}

func getSuffix(matter FrontMatter, path string) string {
	if matter.Title != "" {
		return matter.Title
	}
	filename := filepath.Base(path)
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

// getPrefix returns the first category or tag that is not generic.
// It might return an empty string if no category or tag is found.
func getPrefix(matter FrontMatter) string {
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
	s = strings.Replace(s, " ", "-", -1)
	s = strings.Replace(s, "_", "-", -1)
	// Remove all non-alphanumeric characters
	s = regexp.MustCompile("[^a-z0-9-]").ReplaceAllString(s, "")
	// Remove successive hyphens
	s = regexp.MustCompile("-+").ReplaceAllString(s, "-")
	return s
}
