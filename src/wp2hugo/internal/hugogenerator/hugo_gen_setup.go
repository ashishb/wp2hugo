package hugogenerator

import (
	"fmt"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugogenerator/hugopage"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/rs/zerolog/log"
	"io"
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
	fontName         string
	imageURLProvider hugopage.ImageURLProvider
	outputDirPath    string
	downloadMedia    bool
	wpInfo           wpparser.WebsiteInfo
	mediaProvider    MediaProvider
}

type MediaProvider interface {
	GetReader(url string) (io.Reader, error)
}

func NewGenerator(outputDirPath string, downloadMedia bool, fontName string,
	mediaProvider MediaProvider, info wpparser.WebsiteInfo) *Generator {
	return &Generator{
		fontName:         fontName,
		imageURLProvider: newImageURLProvider(info),
		outputDirPath:    outputDirPath,
		mediaProvider:    mediaProvider,
		downloadMedia:    downloadMedia,
		wpInfo:           info,
	}
}

func (g Generator) Generate() error {
	info := g.wpInfo
	siteDir, err := g.setupHugo(g.outputDirPath)
	if err != nil {
		return err
	}
	if err = updateConfig(*siteDir, info); err != nil {
		return err
	}
	if err = g.writePages(*siteDir, info); err != nil {
		return err
	}
	if err = g.writePosts(*siteDir, info); err != nil {
		return err
	}
	if err = setupArchivePage(*siteDir); err != nil {
		return err
	}
	if err = setupSearchPage(*siteDir); err != nil {
		return err
	}
	if err = setupFont(*siteDir, g.fontName); err != nil {
		return err
	}
	if err = WriteCustomShortCodes(*siteDir); err != nil {
		return err
	}

	if err = setupRssFeedFormat(*siteDir); err != nil {
		return err
	}

	if g.downloadMedia {
		url1 := info.Link + "/favicon.ico"
		media, err := g.mediaProvider.GetReader(url1)
		if err != nil {
			return fmt.Errorf("error fetching media file %s: %s", url1, err)
		}
		if err = writeFavicon(path.Join(*siteDir, "static"), media); err != nil {
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
		"git clone https://github.com/adityatelange/hugo-PaperMod themes/PaperMod --depth=1",
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

func (g Generator) writePages(outputDirPath string, info wpparser.WebsiteInfo) error {
	if len(info.Pages) == 0 {
		log.Info().Msg("No pages to write")
		return nil
	}

	pagesDir := path.Join(outputDirPath, "content", "pages")
	if err := utils.CreateDirIfNotExist(pagesDir); err != nil {
		return err
	}

	// Write pages
	for _, page := range info.Pages {
		pagePath := path.Join(pagesDir, fmt.Sprintf("%s.md", page.Filename()))
		if err := g.writePage(outputDirPath, pagePath, page.CommonFields); err != nil {
			return err
		}
	}

	return nil
}

func (g Generator) writePosts(outputDirPath string, info wpparser.WebsiteInfo) error {
	if len(info.Posts) == 0 {
		log.Info().Msg("No posts to write")
		return nil
	}

	postsDir := path.Join(outputDirPath, "content", "posts")
	if err := utils.CreateDirIfNotExist(postsDir); err != nil {
		return err
	}

	// Write posts
	for _, post := range info.Posts {
		postPath := path.Join(postsDir, fmt.Sprintf("%s.md", post.Filename()))
		if err := g.writePage(outputDirPath, postPath, post.CommonFields); err != nil {
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

func (g Generator) writePage(outputMediaDirPath string, pagePath string, page wpparser.CommonFields) error {
	w, err := os.OpenFile(pagePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening page file: %s", err)
	}
	defer w.Close()

	pageURL, err := url.Parse(page.Link)
	if err != nil {
		return fmt.Errorf("error parsing page URL: %s", err)
	}

	p, err := hugopage.NewPage(
		g.imageURLProvider,
		*pageURL, page.Title, page.PublishDate,
		page.PublishStatus == wpparser.PublishStatusDraft || page.PublishStatus == wpparser.PublishStatusPending,
		page.Categories, page.Tags, page.Content, page.GUID)
	if err != nil {
		return fmt.Errorf("error creating Hugo page: %s", err)
	}
	if err = p.Write(w); err != nil {
		return err
	}
	log.Info().Msgf("Page written: %s", pagePath)

	links := p.WPImageLinks()
	log.Debug().
		Str("page", page.Title).
		Int("links", len(links)).
		Msgf("Embedded media links")
	prefixes := make([]string, 0)
	pageURL.Host = strings.TrimPrefix(pageURL.Host, "www.")
	prefixes = append(prefixes, fmt.Sprintf("https://%s", pageURL.Host))
	prefixes = append(prefixes, fmt.Sprintf("http://%s", pageURL.Host))
	prefixes = append(prefixes, fmt.Sprintf("https://www.%s", pageURL.Host))
	prefixes = append(prefixes, fmt.Sprintf("http://www.%s", pageURL.Host))

	if g.downloadMedia {
		log.Debug().
			Int("links", len(links)).
			Msg("Downloading media files")
		for _, link := range links {
			for _, prefix := range prefixes {
				link = strings.TrimPrefix(link, prefix)
			}
			if !strings.HasPrefix(link, "/") {
				log.Warn().
					Str("link", link).
					Str("source", page.Link).
					Msg("non-relative link")
			}
			outputFilePath := fmt.Sprintf("%s/static/%s", outputMediaDirPath, strings.TrimSuffix(link, "/"))
			if !strings.HasPrefix(link, "http") {
				link = "https://ashishb.net/" + link
			}
			media, err := g.mediaProvider.GetReader(link)
			if err != nil {
				return fmt.Errorf("error fetching media file %s: %s", link, err)
			}
			if err = download(outputFilePath, media); err != nil {
				return fmt.Errorf("error downloading media file: %s embedded in %s", err, page.Link)
			}
		}
	}
	return nil
}
