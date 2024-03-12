package hugogenerator

import (
	"fmt"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/rs/zerolog/log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

const _archiveContent = `
---
title: "All"
layout: "archives"
url: "/all/"
summary: archives
---
`

const _searchContent = `
---
title: "Search" # in any language you want
layout: "search" # necessary for search
summary: "Search"
url: "/search/"
placeholder: "placeholder text in search input box"
---
`

type Generator struct {
}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g Generator) Generate(info wpparser.WebsiteInfo, mediaSourceURL string, outputDirPath string) error {
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
	if err = writePosts(*siteDir, info); err != nil {
		return err
	}
	if err = setupArchivePage(*siteDir); err != nil {
		return err
	}
	if err = setupSearchPage(*siteDir); err != nil {
		return err
	}
	if err = writeCustomShortCodes(*siteDir); err != nil {
		return err
	}

	if err = setupRssFeedFormat(*siteDir); err != nil {
		return err
	}

	if mediaSourceURL != "" {
		if err = writeFavicon(*siteDir, mediaSourceURL); err != nil {
			return err
		}
	}
	log.Debug().
		Str("cmd", fmt.Sprintf("cd %s && hugo serve", *siteDir)).
		Msg("Hugo site has been generated")
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
		"git submodule add --depth=1 https://github.com/adityatelange/hugo-PaperMod.git themes/PaperMod",
		"git submodule update --init --recursive",
		`echo "theme: 'PaperMod'">> hugo.yaml`,
		// Verify that the site is setup correctly
		"hugo",
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

func writePages(outputDirPath string, info wpparser.WebsiteInfo) error {
	if len(info.Pages) == 0 {
		log.Info().Msg("No pages to write")
		return nil
	}

	pagesDir := path.Join(outputDirPath, "content", "pages")
	if err := createDirIfNotExist(pagesDir); err != nil {
		return err
	}

	// Write pages
	for _, page := range info.Pages {
		pagePath := path.Join(pagesDir, fmt.Sprintf("%s.md", page.Filename()))
		if err := writePage(pagePath, page.CommonFields); err != nil {
			return err
		}
	}

	return nil
}

func writePosts(outputDirPath string, info wpparser.WebsiteInfo) error {
	if len(info.Posts) == 0 {
		log.Info().Msg("No posts to write")
		return nil
	}

	postsDir := path.Join(outputDirPath, "content", "posts")
	if err := createDirIfNotExist(postsDir); err != nil {
		return err
	}

	// Write posts
	for _, post := range info.Posts {
		postPath := path.Join(postsDir, fmt.Sprintf("%s.md", post.Filename()))
		if err := writePage(postPath, post.CommonFields); err != nil {
			return err
		}
	}
	return nil
}

// Ref: https://adityatelange.github.io/hugo-PaperMod/posts/papermod/papermod-features/#archives-layout
func setupArchivePage(siteDir string) error {
	filePath := path.Join(siteDir, "content", "archives.md")
	content := _archiveContent
	return writeFile(filePath, []byte(content))
}

// Ref: https://adityatelange.github.io/hugo-PaperMod/posts/papermod/papermod-features/#search-page
func setupSearchPage(siteDir string) error {
	filePath := path.Join(siteDir, "content", "search.md")
	content := _searchContent
	return writeFile(filePath, []byte(content))
}

func writeFavicon(outputDirPath string, websiteURL string) error {
	log.Debug().Msg("Fetching and writing favicon")
	if err := createDirIfNotExist(path.Join(outputDirPath, "static")); err != nil {
		return err
	}
	filePath := path.Join(outputDirPath, "static", "favicon.ico")
	if !strings.HasPrefix(websiteURL, "http") {
		websiteURL = "https://" + websiteURL
	}
	url1 := fmt.Sprintf("%s/favicon.ico", websiteURL)
	resp, err := http.Get(url1)
	if err != nil {
		return fmt.Errorf("error fetching favicon: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error fetching favicon: %s", resp.Status)
	}
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening favicon file: %s", err)
	}
	defer file.Close()
	return resp.Write(file)
}

func writeFile(filePath string, content []byte) error {
	w, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening archive file: %s", err)
	}
	defer w.Close()
	if _, err := w.Write(content); err != nil {
		return fmt.Errorf("error writing to archive file: %s", err)
	}
	return nil
}

func createDirIfNotExist(dirPath string) error {
	if err := os.MkdirAll(dirPath, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("error creating directory: %s", err)
	}
	return nil
}

func writePage(pagePath string, page wpparser.CommonFields) error {
	w, err := os.OpenFile(pagePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening page file: %s", err)
	}
	defer w.Close()

	pageURL, err := url.Parse(page.Link)
	if err != nil {
		return fmt.Errorf("error parsing page URL: %s", err)
	}

	p := _Page{
		AbsoluteURL: *pageURL,
		Title:       page.Title,
		PublishDate: page.PublishDate,
		Draft:       page.PublishStatus == wpparser.PublishStatusDraft || page.PublishStatus == wpparser.PublishStatusPending,
		Categories:  page.Categories,
		Tags:        page.Tags,
		HTMLContent: page.Content,
		GUID:        page.GUID,
	}
	if err = p.Write(w); err != nil {
		return err
	}
	log.Info().Msgf("Page written: %s", pagePath)
	return nil
}
