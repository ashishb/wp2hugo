package hugogenerator

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugogenerator/hugopage"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/nginxgenerator"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/rs/zerolog/log"
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

// Find image media thumbnails resized by WP, like `some-file-1920x1080.jpg`
var _resizedMedia = regexp.MustCompile(`(.*)-\d+x\d+\.(jpg|jpeg|png|webp|gif)`)

type Generator struct {
	fontName         string
	imageURLProvider hugopage.ImageURLProvider
	outputDirPath    string
	wpInfo           wpparser.WebsiteInfo

	// Media related
	mediaProvider                  MediaProvider
	downloadMedia                  bool
	downloadAll                    bool
	continueOnMediaDownloadFailure bool

	// Nginx related
	generateNgnixConfig bool
	ngnixConfig         *nginxgenerator.Config
}

type MediaProvider interface {
	GetReader(url string) (io.Reader, error)
}

func NewGenerator(outputDirPath string, fontName string,
	mediaProvider MediaProvider, downloadMedia bool, downloadAll bool, continueOnMediaDownloadFailure bool,
	generateNgnixConfig bool, info wpparser.WebsiteInfo) *Generator {
	var ngnixConfig *nginxgenerator.Config
	if generateNgnixConfig {
		ngnixConfig = nginxgenerator.NewConfig()
	}
	return &Generator{
		fontName:         fontName,
		imageURLProvider: newImageURLProvider(info),
		outputDirPath:    outputDirPath,
		wpInfo:           info,

		// Media related
		mediaProvider:                  mediaProvider,
		downloadMedia:                  downloadMedia,
		downloadAll:                    downloadAll,
		continueOnMediaDownloadFailure: continueOnMediaDownloadFailure,

		// Nginx related
		generateNgnixConfig: generateNgnixConfig,
		ngnixConfig:         ngnixConfig,
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

	if g.downloadAll {
		if err = g.downloadAllMedia(*siteDir, info); err != nil {
			return err
		}
	}

	if err = g.writePages(*siteDir, info); err != nil {
		return err
	}
	if err = g.writePosts(*siteDir, info); err != nil {
		return err
	}
	if err = g.writeCustomPosts(*siteDir, info); err != nil {
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
		url1 := info.Link().Scheme + "://" + info.Link().Host + "/favicon.ico"
		media, err := g.mediaProvider.GetReader(url1)
		if err != nil {
			log.Error().
				Err(err).
				Str("url", url1).
				Msg("error fetching favicon")
		} else {
			if err = writeFavicon(path.Join(*siteDir, "static"), media); err != nil {
				return err
			}
		}
	}

	if g.generateNgnixConfig {
		nginxConfigPath := path.Join(*siteDir, "nginx.conf")
		if err = os.WriteFile(nginxConfigPath, []byte(g.ngnixConfig.Generate()), 0644); err != nil {
			return err
		} else {
			log.Info().
				Str("nginxConfigPath", nginxConfigPath).
				Msg("Nginx config generated")
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
	// Verify hugo is present
	_, err := exec.LookPath("hugo")
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Hugo not found, install it from https://gohugo.io/")
		return nil, fmt.Errorf("hugo not found, install it from https://gohugo.io/: %s", err)
	}

	// Create output directory
	err = os.MkdirAll(outputDirPath, 0700)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("outputDirPath", outputDirPath).
			Msg("error creating output directory")
		return nil, fmt.Errorf("error creating output directory '%s': %s", outputDirPath, err)
	}

	commands := []string{
		"git version",
		"hugo version",
		// Use YAML file as it is easier to edit it afterward than TOML
		fmt.Sprintf("cd %s && hugo new site %s --format yaml", outputDirPath, siteName),
		fmt.Sprintf("cd %s && git clone https://github.com/adityatelange/hugo-PaperMod themes/PaperMod --depth=1",
			path.Join(outputDirPath, siteName)),
		// Set theme to PaperMod
		fmt.Sprintf(`echo theme: 'PaperMod'>> %s/hugo.yaml`, path.Join(outputDirPath, siteName)),
		// Verify that the site is set up correctly
		fmt.Sprintf("cd %s && hugo", path.Join(outputDirPath, siteName)),
	}
	for i, command := range commands {
		log.Debug().
			Int("step", i+1).
			Int("totalSteps", len(commands)).
			Str("cmd", command).
			Msg("Running Hugo setup command")
		var (
			output []byte
			err    error
		)
		if runtime.GOOS == "windows" {
			output, err = exec.Command("cmd", "/C", command).Output()
		} else {
			// mac & Linux
			output, err = exec.Command("bash", "-c", command).Output()
		}
		if err != nil {
			log.Error().
				Err(err).
				Bytes("output", output).
				Str("cmd", command).
				Msg("error running Hugo setup command")
			return nil, fmt.Errorf("error running Hugo setup command '%s' -> %s", command, err)
		}
		log.Debug().Msgf("Hugo setup output: %s", output)
	}

	// Delete .git directory
	deleteDirs := []string{
		path.Join(outputDirPath, siteName, ".git"),
		path.Join(outputDirPath, siteName, "themes/PaperMod/.git"),
	}
	for _, dir := range deleteDirs {
		err := os.RemoveAll(dir)
		if err != nil {
			log.Error().
				Err(err).
				Str("dir", dir).
				Msg("error removing directory")
			return nil, fmt.Errorf("error removing directory '%s': %s", dir, err)
		}
	}

	siteDir := path.Join(outputDirPath, siteName)
	log.Info().
		Str("location", siteDir).
		Msgf("Hugo site skeleton has been setup")
	return &siteDir, nil
}

func (g Generator) downloadAllMedia(outputDirPath string, info wpparser.WebsiteInfo) error {
	hostname := info.Link().Host
	prefixes := make([]string, 0)
	hostname = strings.TrimPrefix(hostname, "www.")
	prefixes = append(prefixes, fmt.Sprintf("https://%s", hostname))
	prefixes = append(prefixes, fmt.Sprintf("http://%s", hostname))
	prefixes = append(prefixes, fmt.Sprintf("https://www.%s", hostname))
	prefixes = append(prefixes, fmt.Sprintf("http://www.%s", hostname))

	for _, attachment := range info.Attachments() {
		if _, err := downloadMedia(*attachment.GetAttachmentURL(), outputDirPath, prefixes, g, info.Link()); err != nil {
			return err
		}
	}

	return nil
}

func (g Generator) writePages(outputDirPath string, info wpparser.WebsiteInfo) error {
	if len(info.Pages()) == 0 {
		log.Info().Msg("No pages to write")
		return nil
	}

	pagesDir := path.Join(outputDirPath, "content", "pages")
	if err := utils.CreateDirIfNotExist(pagesDir); err != nil {
		return err
	}

	// Write pages
	for _, page := range info.Pages() {
		pagePath := getFilePath(pagesDir, page.Filename())
		if err := g.writePage(outputDirPath, pagePath, page.CommonFields); err != nil {
			return err
		}
		// Redirect from old URL to new URL
		g.maybeAddNginxRedirect(page.CommonFields)
	}

	return nil
}

func (g Generator) writeCustomPosts(outputDirPath string, info wpparser.WebsiteInfo) error {
	if len(info.CustomPosts()) == 0 {
		log.Info().Msg("No custom posts to write")
		return nil
	}

	// Write custom posts
	for _, page := range info.CustomPosts() {
		// Dynamically handle post type for target folder
		pagesDir := path.Join(outputDirPath, "content", *page.PostType)
		if err := utils.CreateDirIfNotExist(pagesDir); err != nil {
			return err
		}

		pagePath := getFilePath(pagesDir, page.Filename())
		if err := g.writePage(outputDirPath, pagePath, page.CommonFields); err != nil {
			return err
		}
		// Redirect from old URL to new URL
		g.maybeAddNginxRedirect(page.CommonFields)
	}

	return nil
}

func (g Generator) maybeAddNginxRedirect(page wpparser.CommonFields) {
	if !g.generateNgnixConfig {
		return
	}

	if page.GUID.Value == "" {
		return
	}

	u1, err := url.Parse(strings.TrimSpace(page.GUID.Value))
	if err != nil {
		log.Warn().
			Err(err).
			Str("url", page.GUID.Value).
			Msg("error parsing GUID as URL")
		return
	}

	u2, err := url.Parse(strings.TrimSpace(page.Link))
	if err != nil {
		log.Warn().
			Err(err).
			Str("url", page.Link).
			Msg("error parsing link as URL")
		return
	}

	if !sameHost(*u1, *u2) {
		return
	}

	oldURLPathWithQuery := u1.Path + "?" + u1.RawQuery
	newPath := u2.Path
	if err := g.ngnixConfig.AddRedirect(oldURLPathWithQuery, newPath); err != nil {
		log.Warn().
			Err(err).
			Str("oldURL", oldURLPathWithQuery).
			Str("newURL", page.Link).
			Msg("error adding nginx redirect")
		return
	}
}

func sameHost(url1 url.URL, url2 url.URL) bool {
	return strings.TrimSuffix(url1.Host, "/") == strings.TrimSuffix(url2.Host, "/")
}

// Sometimes multiple pages have the same filename
// Ref: https://github.com/ashishb/wp2hugo/issues/7
func getFilePath(pagesDir string, baseFileName string) string {
	pagePath := path.Join(pagesDir, fmt.Sprintf("%s.md", baseFileName))
	if utils.FileExists(pagePath) {
		for i := 1; ; i++ {
			log.Info().
				Str("baseFileName", baseFileName).
				Str("pagePath", pagePath).
				Msg("File already exists, trying another filename")
			pagePath = path.Join(pagesDir, fmt.Sprintf("%s-%d.md", baseFileName, i))
			if !utils.FileExists(pagePath) {
				break
			}
		}
	}
	return pagePath
}

func (g Generator) writePosts(outputDirPath string, info wpparser.WebsiteInfo) error {
	if len(info.Posts()) == 0 {
		log.Info().Msg("No posts to write")
		return nil
	}

	postsDir := path.Join(outputDirPath, "content", "posts")
	if err := utils.CreateDirIfNotExist(postsDir); err != nil {
		return err
	}

	// Write posts
	for _, post := range info.Posts() {
		postPath := getFilePath(postsDir, post.Filename())
		if err := g.writePage(outputDirPath, postPath, post.CommonFields); err != nil {
			return err
		}
		// Redirect from old URL to new URL
		g.maybeAddNginxRedirect(post.CommonFields)
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
	if _, err := w.Write(content); err != nil {
		return fmt.Errorf("error writing to archive file: %s", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("error closing archive file: %s", err)
	}
	return nil
}

func (g Generator) writePage(outputMediaDirPath string, pagePath string, page wpparser.CommonFields) error {
	pageURL, err := url.Parse(page.Link)
	if err != nil {
		return fmt.Errorf("error parsing page URL: %s", err)
	}

	p, err := g.newHugoPage(pageURL, page)
	if err != nil {
		return fmt.Errorf("error creating Hugo page: %s", err)
	}

	if g.downloadMedia {
		urlReplacements, err := g.downloadPageMedia(outputMediaDirPath, p, pageURL)
		if err != nil {
			return err
		} else {
			p.Replace(urlReplacements)
		}
	}

	w, err := os.OpenFile(pagePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening page file: %s", err)
	}

	if err = p.Write(w); err != nil {
		return fmt.Errorf("error writing page file: %s", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("error closing page file: %s", err)
	}

	log.Info().Msgf("Page written: %s", pagePath)

	return nil
}

func (g Generator) newHugoPage(pageURL *url.URL, page wpparser.CommonFields) (*hugopage.Page, error) {
	return hugopage.NewPage(
		g.imageURLProvider,
		*pageURL, page.Author, page.Title, page.PublishDate,
		page.PublishStatus == wpparser.PublishStatusDraft || page.PublishStatus == wpparser.PublishStatusPending,
		page.Categories, page.Tags, g.wpInfo.GetAttachmentsForPost(page.PostID),
		page.Footnotes, page.Content, page.GUID, page.FeaturedImageID, page.PostFormat, page.CustomMetaData, page.Taxonomies)
}

func downloadMedia(link string, outputMediaDirPath string, prefixes []string, g Generator, pageURL *url.URL) (map[string]string, error) {

	// Uniformize protocol-less links: add protocol
	if strings.HasPrefix(link, "//") {
		link = strings.Replace(link, "//", fmt.Sprintf("%s://", pageURL.Scheme), 1)
	}

	// Turn all absolute links pointing to current host into relative links
	for _, prefix := range prefixes {
		link = strings.TrimPrefix(link, prefix)
	}

	// Now, all absolute links point to external domains:
	// bypass
	if !strings.HasPrefix(link, "/") {
		log.Warn().
			Str("link", link).
			Str("source", pageURL.String()).
			Msg("non-relative link (skipped for download)")
		return nil, nil
	}

	relativeLink := link
	outputFilePath := fmt.Sprintf("%s/static/%s", outputMediaDirPath,
		strings.TrimSuffix(strings.Split(link, "?")[0], "/"))

	if strings.HasPrefix(link, "http") {
		// do nothing in case of absolute URL
		// this case should not happen, the function would have returned already
	} else if strings.HasPrefix(link, "/") {
		// relative URL to the base of the website
		// turn it to absolute URL
		link = g.wpInfo.Link().Scheme + "://" + g.wpInfo.Link().Host + link
	} else {
		link = strings.TrimSuffix(g.wpInfo.Link().String(), "/") + "/" + link
	}

	// Try full-res images first.
	// It is assumed here that Hugo will handle responsive sizes and such internally.
	// see https://discourse.gohugo.io/t/hugo-image-processing-and-responsive-images/43110/4
	fullResLink := _resizedMedia.ReplaceAllString(link, "$1.$2")
	media, err := g.mediaProvider.GetReader(fullResLink)

	urlReplacement := make(map[string]string)

	if err != nil {
		// If full-res image not found, try again with resized one.
		if strings.Compare(fullResLink, link) != 0 {
			log.Info().
				Str("fullResLink", fullResLink).
				Str("link", link).
				Msg("full-resolution image file not found, falling back to resized thumbnail")
			media, err = g.mediaProvider.GetReader(link)
		} else {
			new_link := _resizedMedia.ReplaceAllString(relativeLink, "$1.$2")
			urlReplacement[relativeLink] = new_link
			urlReplacement[link] = new_link
		}
	} else {
		// If full-res image found, update target file path too
		if strings.Compare(fullResLink, link) != 0 {
			outputFilePath = _resizedMedia.ReplaceAllString(outputFilePath, "$1.$2")
			log.Info().
				Str("fullResLink", fullResLink).
				Str("link", link).
				Msg("resized thumbnail was replaced by full-resolution image")

			new_link := _resizedMedia.ReplaceAllString(relativeLink, "$1.$2")
			urlReplacement[relativeLink] = new_link
			urlReplacement[link] = new_link
		}
	}

	// Note: we will substitute resized image links with full-res image links
	// after all links are turned to relative in the generated Markdown.
	// Thus we register URL replacements as relative links.

	if err != nil {
		if g.continueOnMediaDownloadFailure {
			log.Error().
				Err(err).
				Str("mediaLink", link).
				Str("pageLink", pageURL.String()).
				Str("outputFilePath", outputFilePath).
				Msg("error fetching media file")
			return urlReplacement, nil
		} else {
			return nil, fmt.Errorf("error fetching media file %s: %s", link, err)
		}
	}

	if err = download(outputFilePath, media); err != nil {
		if g.continueOnMediaDownloadFailure {
			log.Error().
				Err(err).
				Str("mediaLink", link).
				Str("pageLink", pageURL.String()).
				Msg("error downloading media file")
		} else {
			return nil, fmt.Errorf("error downloading media file: %s embedded in %s", err, pageURL.String())
		}
	}

	return urlReplacement, nil
}

func (g Generator) downloadPageMedia(outputMediaDirPath string, p *hugopage.Page, pageURL *url.URL) (map[string]string, error) {
	links := p.WPMediaLinks()
	log.Debug().
		Str("page", pageURL.String()).
		Int("links", len(links)).
		Msgf("Embedded media links")
	log.Debug().
		Int("links", len(links)).
		Strs("links", links).
		Msg("Downloading media files")

	hostname := pageURL.Host
	prefixes := make([]string, 0)
	hostname = strings.TrimPrefix(hostname, "www.")
	prefixes = append(prefixes, fmt.Sprintf("https://%s", hostname))
	prefixes = append(prefixes, fmt.Sprintf("http://%s", hostname))
	prefixes = append(prefixes, fmt.Sprintf("https://www.%s", hostname))
	prefixes = append(prefixes, fmt.Sprintf("http://www.%s", hostname))

	urlReplacements := make(map[string]string)

	for _, link := range links {
		if replacement, err := downloadMedia(link, outputMediaDirPath, prefixes, g, pageURL); err != nil {
			return nil, err
		} else {
			for k, v := range replacement {
				urlReplacements[k] = v
			}
		}
	}
	return urlReplacements, nil
}
