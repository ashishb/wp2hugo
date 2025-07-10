package frontmatterhelper

import (
	"os"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
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

	Layout      *string  `yaml:"layout,omitempty"`      // Used by Hugo papermod theme
	Placeholder *string  `yaml:"placeholder,omitempty"` // Used by Hugo papermod theme for search page
	GUID        string   `yaml:"guid"`                  // For RSS and Atom feeds
	Author      []string `yaml:"author"`
	Cover       struct {
		Alt     *string `yaml:"alt,omitempty"`
		Caption *string `yaml:"caption,omitempty"`
		Image   string  `yaml:"image,omitempty"`
	} `yaml:"cover"`
	Aliases []string `yaml:"aliases"`           // For redirects
	ShowToc *bool    `yaml:"ShowToc,omitempty"` // For Hugo papermod theme
	TocOpen *bool    `yaml:"TocOpen,omitempty"` // For Hugo papermod theme
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

// GetSelectiveFrontMatter reads the front matter from a file and returns it as a FrontMatter struct.
func GetSelectiveFrontMatter(path string) (*FrontMatter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	var matter FrontMatter
	if false { // Set this to true for strict parsing and this catches typos in the front matter
		// Ref: https://github.com/adrg/frontmatter/issues/50
		formats := []*frontmatter.Format{
			frontmatter.NewFormat("---", "---", yaml.UnmarshalStrict),
		}
		_, err = frontmatter.MustParse(file, &matter, formats...)
	} else {
		_, err = frontmatter.Parse(file, &matter)
	}
	if err != nil {
		log.Error().
			Err(err).
			Str("path", path).
			Msg("Error parsing front matter")
		return nil, err
	}
	return &matter, nil
}
