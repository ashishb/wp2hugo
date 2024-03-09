package hugogenerator

import (
	"fmt"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

type Generator struct {
}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g Generator) Generate(info *wpparser.WebsiteInfo, outputDirPath string) error {
	siteDir, err := g.setupHugo(outputDirPath)
	if err != nil {
		return err
	}
	if err = updateConfig(*siteDir, info); err != nil {
		return err
	}
	if err = writePages(*siteDir, info); err != nil {
		return err
	}
	return nil
}

func (g Generator) setupHugo(outputDirPath string) (*string, error) {
	// Replace spaces and colons with dashes
	timeFormat := time.Now().Format(
		strings.ReplaceAll(strings.ReplaceAll(time.DateTime, " ", "-"), ":", "-"))
	siteName := fmt.Sprintf("generated-%s", timeFormat)
	log.Debug().
		Str("siteName", siteName).
		Msg("Setting up Hugo site")
	commands := []string{
		"hugo version",
		"cd " + outputDirPath,
		// Use YMAL file as it is easier to edit it afterwards than TOML
		fmt.Sprintf("hugo new site %s --format yaml", siteName),
		"cd " + siteName,
		"git init",
		"git submodule add https://github.com/theNewDynamic/gohugo-theme-ananke.git themes/ananke",
		`echo "theme: 'ananke'" >> hugo.yaml`,
	}
	combinedCommand := strings.Join(commands, " && ")
	log.Debug().Msg("Running Hugo setup commands")
	output, err := exec.Command("bash", "-c", combinedCommand).Output()
	if err != nil {
		return nil, fmt.Errorf("error running Hugo setup commands: %s", err)
	}
	log.Debug().Msgf("Hugo setup output: %s", output)
	siteDir := path.Join(outputDirPath, siteName)
	log.Info().
		Str("location", siteDir).
		Msgf("Hugo site skeleton has been setup")
	return &siteDir, nil
}

func updateConfig(siteDir string, info *wpparser.WebsiteInfo) error {
	configPath := path.Join(siteDir, "hugo.yaml")
	r, err := os.OpenFile(configPath, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening config file: %s", err)
	}
	defer r.Close()

	type Config struct {
		BaseURL      string `yaml:"baseURL"`
		LanguageCode string `yaml:"languageCode"`
		Title        string `yaml:"title"`
		Theme        string `yaml:"theme"`
		// These will be used for OpenGraph information
		Params struct {
			Description string `yaml:"description"`
		} `yaml:"params"`
	}

	var config Config
	if err := yaml.NewDecoder(r).Decode(&config); err != nil {
		return fmt.Errorf("error unmarshalling config: %s", err)
	}
	config.Title = info.Title
	config.BaseURL = info.Link
	config.LanguageCode = info.Language
	config.Params.Description = info.Description

	if err = r.Close(); err != nil {
		return fmt.Errorf("error closing config file: %s", err)
	}
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling config: %s", err)
	}
	w, err := os.OpenFile(configPath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening config file: %s", err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("error writing to config file: %s", err)
	}
	defer w.Close()
	log.Info().Msgf("Updating config file: %s", configPath)
	return w.Close()
}

func writePages(outputDirPath string, info *wpparser.WebsiteInfo) error {
	// Create content directory
	contentDir := path.Join(outputDirPath, "content")
	if err := os.Mkdir(contentDir, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("error creating content directory: %s", err)
	}

	// Write pages
	for _, page := range info.Pages {
		pagePath := path.Join(contentDir, fmt.Sprintf("%s.md", page.Title))
		if err := writePage(pagePath, page); err != nil {
			return err
		}
	}

	//// Write posts
	//for _, post := range info.Posts {
	//	postPath := path.Join(postsDir, fmt.Sprintf("%s.md", post.Slug))
	//	if err := writePost(postPath, post); err != nil {
	//		return err
	//	}
	//}
	return nil
}

func writePage(pagePath string, page wpparser.PageInfo) error {
	w, err := os.OpenFile(pagePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening page file: %s", err)
	}
	defer w.Close()

	p := _Page{
		Title:       page.Title,
		PublishDate: *page.PublishDate,
		Draft:       page.PublishStatus == wpparser.PublishStatusDraft,
		Categories:  page.Categories,
		Tags:        page.Tags,
		HTMLContent: page.Content,
	}
	if err = p.Write(w); err != nil {
		return err
	}
	log.Info().Msgf("Page written: %s", pagePath)
	return nil
}
