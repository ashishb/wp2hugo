package hugopage

import (
	"fmt"
	"github.com/mmcdole/gofeed/rss"
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

	CategoryName = "category"
	TagName      = "tag"
)

type Page struct {
	// This is the original URL of the page from the WordPress site
	AbsoluteURL url.URL

	Title       string
	PublishDate *time.Time
	Draft       bool
	Categories  []string
	Tags        []string
	GUID        *rss.GUID

	// HTMLContent is the HTML content of the page that will be
	// transformed to Markdown
	HTMLContent string
}

var _markdownImageLinks = regexp.MustCompile(`!\[.*?]\((.+?)\)`)

// Extracts "src" from Hugo figure shortcode
// {{< figure align=aligncenter width=905 src="/wp-content/uploads/2023/01/Stollemeyer-castle-1024x768.jpg" alt="" >}}
var _hugoFigureLinks = regexp.MustCompile(`{{< figure.*?src="(.+?)".*? >}}`)

func (page Page) getRelativeURL() string {
	return page.AbsoluteURL.Path
}

func (page Page) Write(w io.Writer) error {
	if err := page.writeMetadata(w); err != nil {
		return err
	}
	if err := page.writeContent(w); err != nil {
		return err
	}
	return nil
}

func (page Page) WPImageLinks() ([]string, error) {
	markdown, err := page.getMarkdown()
	if err != nil {
		return nil, err
	}
	arr1 := getMarkdownLinks(_markdownImageLinks, *markdown)
	arr2 := getMarkdownLinks(_hugoFigureLinks, *markdown)
	return append(arr1, arr2...), nil
}

func getMarkdownLinks(regex *regexp.Regexp, markdown string) []string {
	var links []string
	matches := regex.FindAllStringSubmatch(markdown, -1)
	for _, match := range matches {
		links = append(links, match[1])
	}
	return links
}

func (page Page) writeMetadata(w io.Writer) error {
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
		metadata[CategoryName] = page.Categories
	}

	if len(page.Tags) > 0 {
		metadata[TagName] = page.Tags
	}
	if page.GUID != nil {
		metadata["GUID"] = page.GUID.Value
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

func (page Page) getMarkdown() (*string, error) {
	if page.HTMLContent == "" {
		return nil, fmt.Errorf("empty HTML content")
	}
	converter := getMarkdownConverter()
	htmlContent := replaceCaptionWithFigure(page.HTMLContent)
	markdown, err := converter.ConvertString(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("error converting HTML to Markdown: %s", err)
	}
	if len(strings.TrimSpace(markdown)) == 0 {
		return nil, fmt.Errorf("empty markdown")
	}
	markdown = ReplaceAbsoluteLinksWithRelative(page.AbsoluteURL.Host, markdown)
	markdown = replaceCatlistWithShortcode(markdown)
	return &markdown, nil
}

func (page Page) writeContent(w io.Writer) error {
	markdown, err := page.getMarkdown()
	if err != nil {
		return err
	}

	if _, err := w.Write([]byte(*markdown)); err != nil {
		return fmt.Errorf("error writing to page file: %s", err)
	}
	return nil
}

func ReplaceAbsoluteLinksWithRelative(hostName string, markdownData string) string {
	log.Debug().
		Str("hostName", hostName).
		Msg("Replacing absolute links with relative links")
	re1 := regexp.MustCompile("https://" + hostName + "/")
	re2 := regexp.MustCompile("http://" + hostName + "/")
	markdownData = re1.ReplaceAllString(markdownData, "/")
	markdownData = re2.ReplaceAllString(markdownData, "/")
	return markdownData
}
