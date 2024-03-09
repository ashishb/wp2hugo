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
	}

	var config Config
	if err := yaml.NewDecoder(r).Decode(&config); err != nil {
		return fmt.Errorf("error unmarshalling config: %s", err)
	}
	config.Title = info.Title
	config.BaseURL = info.Link
	config.LanguageCode = info.Language

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
