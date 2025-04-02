package frontmatterhelper

import (
	"github.com/adrg/frontmatter"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"time"
)

type FrontMatter struct {
	URL         string `yaml:"url"`
	PublishDate string `yaml:"date"`

	Description *string `yaml:"description"`
	Summary     string  `yaml:"summary"`
	Title       string  `yaml:"title"`

	Categories []string `yaml:"category"`
	Draft      string   `yaml:"draft"`
	Tags       []string `yaml:"tag"`
}

func (f *FrontMatter) IsDraft() bool {
	return strings.ToLower(f.Draft) == "true"
}

func (f *FrontMatter) IsInFuture() (bool, error) {
	if f.PublishDate == "" {
		return false, nil
	}
	t1, err := time.Parse(time.RFC3339, f.PublishDate)
	if err != nil {
		return false, err
	}
	return t1.After(time.Now()), nil
}

func (f *FrontMatter) HasDescription() bool {
	return f.Description != nil && strings.TrimSpace(*f.Description) != ""
}

func GetSelectiveFrontMatter(path string) (*FrontMatter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	var matter FrontMatter
	_, err = frontmatter.MustParse(file, &matter)
	if err != nil {
		log.Error().
			Err(err).
			Str("path", path).
			Msg("Error parsing front matter")
		return nil, err
	}
	return &matter, nil
}
