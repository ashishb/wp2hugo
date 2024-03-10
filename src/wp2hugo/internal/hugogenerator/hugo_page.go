package hugogenerator

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"io"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	// Seems to be undocumented, but this is the date format used by Hugo
	_hugoDateFormat = "2006-01-02T15:04:05-07:00"

	_CategoryName = "category"
	_TagName      = "tag"
)

type _Page struct {
	// This is the original URL of the page from the WordPress site
	AbsoluteURL url.URL

	Title       string
	PublishDate *time.Time
	Draft       bool
	Categories  []string
	Tags        []string

	// HTMLContent is the HTML content of the page that will be
	// transformed to Markdown
	HTMLContent string
}

func (page _Page) getRelativeURL() string {
	return page.AbsoluteURL.Path
}

func (page _Page) Write(w io.Writer) error {
	if err := page.writeMetadata(w); err != nil {
		return err
	}
	if err := page.writeContent(w); err != nil {
		return err
	}
	return nil
}

func (page _Page) writeMetadata(w io.Writer) error {
	metadata := make(map[string]any)
	metadata["url"] = page.getRelativeURL()
	metadata["title"] = page.Title
	if page.PublishDate != nil {
		metadata["date"] = page.PublishDate.Format(_hugoDateFormat)
	}
	if page.Draft {
		metadata["draft"] = "true"
	}

	if len(page.Categories) > 0 {
		metadata[_CategoryName] = page.Categories
	}

	if len(page.Tags) > 0 {
		metadata[_TagName] = page.Tags
	}

	combinedMetadata, err := yaml.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("error marshalling metadata: %s", err)
	}
	combinedMetadataStr := fmt.Sprintf("---\n%s\n---\n", string(combinedMetadata))
	if _, err := w.Write([]byte(combinedMetadataStr)); err != nil {
		return fmt.Errorf("error writing to page file: %s", err)
	}
	return nil
}

func (page _Page) writeContent(w io.Writer) error {
	if page.HTMLContent == "" {
		return fmt.Errorf("empty HTML content")
	}
	converter := getMarkdownConverter()
	markdown, err := converter.ConvertString(page.HTMLContent)
	if err != nil {
		return fmt.Errorf("error converting HTML to Markdown: %s", err)
	}
	if len(strings.TrimSpace(markdown)) == 0 {
		return fmt.Errorf("empty markdown")
	}
	markdown = replaceAbsoluteLinksWithRelative(page.AbsoluteURL.Host, markdown)

	if _, err := w.Write([]byte(markdown)); err != nil {
		return fmt.Errorf("error writing to page file: %s", err)
	}
	return nil
}

func replaceAbsoluteLinksWithRelative(hostName string, markdownData string) string {
	log.Debug().
		Str("hostName", hostName).
		Msg("Replacing absolute links with relative links")
	re1 := regexp.MustCompile("https://" + hostName + "/")
	re2 := regexp.MustCompile("http://" + hostName + "/")
	markdownData = re1.ReplaceAllString(markdownData, "/")
	markdownData = re2.ReplaceAllString(markdownData, "/")
	return markdownData
}
