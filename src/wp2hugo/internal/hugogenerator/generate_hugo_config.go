package hugogenerator

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugogenerator/hugopage"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type _HugoNavMenu struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
	// Weight is the order in which the menu items will be displayed
	Weight int `yaml:"weight"`
}

type _HugoAttachment struct {
	Path  string    `yaml:"path"`
	Title string    `yaml:"title"`
	ID    string    `yaml:"id"`
	Date  time.Time `yaml:"published"`
}

type _HugoConfig struct {
	BaseURL      string `yaml:"baseURL"`
	LanguageCode string `yaml:"languageCode"`
	Title        string `yaml:"title"`
	Theme        string `yaml:"theme"`
	Taxonomies   struct {
		Category string `yaml:"category"`
		Tag      string `yaml:"tag"`
	}
	// These will be used for OpenGraph information
	Params struct {
		Description         string `yaml:"description"`
		DefaultTheme        string `yaml:"defaultTheme"`
		DisableThemeToggle  bool   `yaml:"disableThemeToggle"`
		ShowShareButtons    bool   `yaml:"showShareButtons"`
		ShowReadingTime     bool   `yaml:"showReadingTime"`
		ShowToc             bool   `yaml:"showToc"`
		ShowBreadCrumbs     bool   `yaml:"showBreadCrumbs"`
		ShowCodeCopyButtons bool   `yaml:"showCodeCopyButtons"`
		Comments            bool   `yaml:"comments"`
		HideFooter          bool   `yaml:"hideFooter"`
		Assets              struct {
			Favicon     string `yaml:"favicon"`
			DisableHLJS bool   `yaml:"disableHLJS"`
		} `yaml:"assets"`
	} `yaml:"params"`
	Markup struct {
		Highlight struct {
			CodeFences  bool   `yaml:"codeFences"`
			GuessSyntax bool   `yaml:"guessSyntax"`
			Style       string `yaml:"style"`
		} `yaml:"highlight"`
		Goldmark struct {
			// Unsafe HTML is needed to nest shortcodes within each other, aka figure inside gallery
			Renderer struct {
				Unsafe bool `yaml:"unsafe"`
			} `yaml:"renderer"`
		} `yaml:"goldmark"`
	}
	Outputs struct {
		Home []string `yaml:"home"`
	}
	OutputFormats struct {
		RSS struct {
			MediaType string `yaml:"mediaType"`
			BaseName  string `yaml:"baseName"`
		} `yaml:"RSS"`
	} `yaml:"outputFormats"`
	Menu struct {
		Main []_HugoNavMenu `yaml:"main"`
	} `yaml:"menu"`
}

func setupLibraryData(siteDir string, info wpparser.WebsiteInfo) error {
	dataPath := path.Join(siteDir, "data", "library.yaml")
	dataDir := path.Dir(dataPath)

	// Create the directory if it doesn't exist
	err := os.MkdirAll(dataDir, 0o755)
	if err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Create the YAML file
	r, err := os.Create(dataPath)
	if err != nil {
		return fmt.Errorf("error creating data file: %w", err)
	}
	defer func() {
		_ = r.Close()
	}()

	// Write the WP media library into data
	library := make([]_HugoAttachment, 0, len(info.Attachments()))
	for _, attachment := range info.Attachments() {
		library = append(library, _HugoAttachment{
			Path:  hugopage.ReplaceAbsoluteLinksWithRelative(info.Link().Host, *attachment.GetAttachmentURL()),
			ID:    attachment.PostID,
			Title: attachment.Title,
			Date:  *attachment.PublishDate,
		})
	}

	data, err := utils.GetYAML(library)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}

	log.Info().Msgf("Updating config file: %s", dataPath)
	return writeFile(dataPath, data)
}

func updateConfig(siteDir string, info wpparser.WebsiteInfo) error {
	configPath := path.Join(siteDir, "hugo.yaml")
	r, err := os.OpenFile(configPath, os.O_RDONLY, 0o644)
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}
	defer func() {
		_ = r.Close()
	}()

	var config _HugoConfig
	if err := yaml.NewDecoder(r).Decode(&config); err != nil {
		return fmt.Errorf("error unmarshalling config: %w", err)
	}
	if config.Theme == "" {
		return errors.New("error: theme is not set in the config file")
	}
	// Ref: https://adityatelange.github.io/hugo-PaperMod/posts/papermod/papermod-faq/
	config.Title = info.Title()
	config.BaseURL = info.Link().String()
	config.LanguageCode = info.Language()
	config.Taxonomies.Category = hugopage.CategoryName
	config.Taxonomies.Tag = hugopage.TagName
	config.Params.Description = info.Description
	config.Params.Assets.Favicon = "/favicon.ico"
	config.Params.Assets.DisableHLJS = true
	// To switch between dark or light according to browser theme
	config.Params.DefaultTheme = "auto"
	config.Params.DisableThemeToggle = true
	config.Params.ShowShareButtons = true
	config.Params.ShowReadingTime = true
	config.Params.ShowToc = false
	config.Params.ShowBreadCrumbs = true
	config.Params.ShowCodeCopyButtons = true
	config.Params.Comments = true
	config.Params.HideFooter = true

	config.Markup.Highlight.CodeFences = true
	config.Markup.Highlight.GuessSyntax = true
	config.Markup.Highlight.Style = "monokai"
	config.Markup.Goldmark.Renderer.Unsafe = true
	// https://adityatelange.github.io/hugo-PaperMod/posts/papermod/papermod-features/#search-page
	config.Outputs.Home = []string{"HTML", "RSS", "JSON"}
	config.OutputFormats.RSS.MediaType = "application/rss+xml"
	// Same as WordPress's feed.xml
	config.OutputFormats.RSS.BaseName = "feed"

	addNavigationLinks(info, &config)
	if err := r.Close(); err != nil {
		return fmt.Errorf("error closing config file: %w", err)
	}
	data, err := utils.GetYAML(config)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}
	log.Info().Msgf("Updating config file: %s", configPath)
	return writeFile(configPath, data)
}

func addNavigationLinks(info wpparser.WebsiteInfo, config *_HugoConfig) {
	if len(info.NavigationLinks()) == 0 {
		return
	}

	searchPresent := false
	for i, link := range info.NavigationLinks() {
		config.Menu.Main = append(config.Menu.Main, _HugoNavMenu{
			Name:   link.Title,
			URL:    hugopage.ReplaceAbsoluteLinksWithRelative(info.Link().Host, link.URL),
			Weight: i + 1,
		})
		if strings.HasSuffix(link.URL, "/search/") {
			searchPresent = true
		}
	}

	// add search at the end of the menu
	if !searchPresent {
		config.Menu.Main = append(config.Menu.Main, _HugoNavMenu{
			Name:   "🔍",
			URL:    "/search/",
			Weight: len(info.NavigationLinks()) + 1,
		})
	}
}
