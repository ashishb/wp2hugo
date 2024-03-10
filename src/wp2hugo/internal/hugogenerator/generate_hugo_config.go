package hugogenerator

import (
	"fmt"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"net/url"
	"os"
	"path"
	"strings"
)

type _HugoNavMenu struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
	// Weight is the order in which the menu items will be displayed
	Weight int `yaml:"weight"`
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
	}
	Outputs struct {
		Home []string `yaml:"home"`
	}
	Menu struct {
		Main []_HugoNavMenu `yaml:"main"`
	} `yaml:"menu"`
}

func updateConfig(siteDir string, info wpparser.WebsiteInfo) error {
	configPath := path.Join(siteDir, "hugo.yaml")
	r, err := os.OpenFile(configPath, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening config file: %s", err)
	}
	defer r.Close()

	var config _HugoConfig
	if err := yaml.NewDecoder(r).Decode(&config); err != nil {
		return fmt.Errorf("error unmarshalling config: %s", err)
	}
	if config.Theme == "" {
		return fmt.Errorf("error: theme is not set in the config file")
	}
	// Ref: https://adityatelange.github.io/hugo-PaperMod/posts/papermod/papermod-faq/
	config.Title = info.Title
	config.BaseURL = info.Link
	config.LanguageCode = info.Language
	config.Taxonomies.Category = _CategoryName
	config.Taxonomies.Tag = _TagName
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
	// https://adityatelange.github.io/hugo-PaperMod/posts/papermod/papermod-features/#search-page
	config.Outputs.Home = []string{"HTML", "RSS", "JSON"}

	if err := addNavigationLinks(info, &config); err != nil {
		return err
	}

	if err := r.Close(); err != nil {
		return fmt.Errorf("error closing config file: %s", err)
	}
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling config: %s", err)
	}
	log.Info().Msgf("Updating config file: %s", configPath)
	return writeFile(configPath, data)
}

func addNavigationLinks(info wpparser.WebsiteInfo, config *_HugoConfig) error {
	if len(info.NavigationLinks) <= 0 {
		return nil
	}
	hostName, err := url.Parse(info.Link)
	if err != nil {
		return fmt.Errorf("error parsing host name: %s", err)
	}

	searchPresent := false

	for i, link := range info.NavigationLinks {
		config.Menu.Main = append(config.Menu.Main, _HugoNavMenu{
			Name:   link.Title,
			URL:    replaceAbsoluteLinksWithRelative(hostName.Host, link.URL),
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
			Weight: len(info.NavigationLinks) + 1,
		})
	}
	return nil
}
