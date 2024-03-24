package hugopage

import (
	"fmt"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/go-enry/go-enry/v2"
	"github.com/mmcdole/gofeed/rss"
	"github.com/rs/zerolog/log"
	"io"
	"net/url"
	"regexp"
	"slices"
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
	absoluteURL url.URL

	metadata map[string]any
	markdown string
}

const _WordPressMoreTag = "<!--more-->"

// In the next step, we will replace this as well
const _customMoreTag = "{{< more >}}"
const _wordPressTocTag = "[toc]"

var (
	_markdownImageLinks = regexp.MustCompile(`!\[.*?]\((.+?)\)`)
	// E.g. <pre class="EnlighterJSRAW" data-enlighter-language="golang">
	_preTagExtractor1 = regexp.MustCompile(`<pre class="EnlighterJSRAW" data-enlighter-language="([^"]+?)".*?>([\s\S]*?)</pre>`)
	// E.g. <pre class="lang:bash" nums="false">
	_preTagExtractor2 = regexp.MustCompile(`<pre class=".*?lang:([^" ]+).*?>([\s\S]*?)</pre>`)
)

// Extracts "src" from Hugo figure shortcode
// {{< figure align=aligncenter width=905 src="/wp-content/uploads/2023/01/Stollemeyer-castle-1024x768.jpg" alt="" >}}
var _hugoFigureLinks = regexp.MustCompile(`{{< figure.*?src="(.+?)".*? >}}`)

// {{< parallaxblur src="/wp-content/uploads/2018/12/bora%5Fbora%5F5%5Fresized.jpg" >}}
var _hugoParallaxBlurLinks = regexp.MustCompile(`{{< parallaxblur.*?src="(.+?)".*? >}}`)

func NewPage(provider ImageURLProvider, pageURL url.URL, title string, publishDate *time.Time, isDraft bool,
	categories []string, tags []string, htmlContent string, guid *rss.GUID) (*Page, error) {
	page := Page{
		absoluteURL: pageURL,
		metadata:    getMetadata(pageURL, title, publishDate, isDraft, categories, tags, guid),
	}
	// htmlContent is the HTML content of the page that will be
	// transformed to Markdown
	markdown, err := page.getMarkdown(provider, htmlContent)
	if err != nil {
		return nil, err
	}
	page.markdown = *markdown
	return &page, nil
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

func (page *Page) WPImageLinks() []string {
	arr1 := getMarkdownLinks(_markdownImageLinks, page.markdown)
	arr2 := getMarkdownLinks(_hugoFigureLinks, page.markdown)
	arr3 := getMarkdownLinks(_hugoParallaxBlurLinks, page.markdown)
	return append(append(arr1, arr2...), arr3...)
}

func getMarkdownLinks(regex *regexp.Regexp, markdown string) []string {
	var links []string
	matches := regex.FindAllStringSubmatch(markdown, -1)
	for _, match := range matches {
		links = append(links, match[1])
	}
	return links
}

func getMetadata(pageURL url.URL, title string, publishDate *time.Time, isDraft bool,
	categories []string, tags []string, guid *rss.GUID) map[string]any {
	metadata := make(map[string]any)
	metadata["url"] = pageURL.Path // Relative URL
	metadata["title"] = title
	if publishDate != nil {
		metadata["date"] = publishDate.Format(_hugoDateFormat)
	}
	if isDraft {
		metadata["draft"] = "true"
	}
	if len(categories) > 0 {
		metadata[CategoryName] = categories
	}
	if len(tags) > 0 {
		metadata[TagName] = tags
	}
	if guid != nil {
		metadata["guid"] = guid.Value
	}
	return metadata
}

func (page *Page) writeMetadata(w io.Writer) error {
	combinedMetadata, err := utils.GetYAML(page.metadata)
	if err != nil {
		return fmt.Errorf("error marshalling metadata: %s", err)
	}
	combinedMetadataStr := fmt.Sprintf("---\n%s\n---\n", string(combinedMetadata))
	if _, err := w.Write([]byte(combinedMetadataStr)); err != nil {
		return fmt.Errorf("error writing to page file: %s", err)
	}
	return nil
}

func (page *Page) getMarkdown(provider ImageURLProvider, htmlContent string) (*string, error) {
	if htmlContent == "" {
		return nil, fmt.Errorf("empty HTML content")
	}
	converter := getMarkdownConverter()
	htmlContent = improvePreTagsWithCode(htmlContent)
	htmlContent = replaceCaptionWithFigure(htmlContent)
	htmlContent = replaceAWBWithParallaxBlur(provider, htmlContent)

	htmlContent = strings.Replace(htmlContent, _WordPressMoreTag, _customMoreTag, 1)
	// This handling is specific to paperMod theme
	// Ref: https://adityatelange.github.io/hugo-PaperMod/posts/papermod/papermod-features/#show-table-of-contents-toc-on-blog-post
	if strings.Contains(htmlContent, _wordPressTocTag) {
		htmlContent = strings.Replace(htmlContent, _wordPressTocTag, "", 1)
		page.metadata["ShowToc"] = true
		page.metadata["TocOpen"] = true
	}
	markdown, err := converter.ConvertString(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("error converting HTML to Markdown: %s", err)
	}
	if len(strings.TrimSpace(markdown)) == 0 {
		return nil, fmt.Errorf("empty markdown")
	}
	if strings.Contains(markdown, _customMoreTag) {
		// Ref: https://gohugo.io/content-management/summaries/#manual-summary-splitting
		page.metadata["summary"] = strings.Split(markdown, _customMoreTag)[0]
		markdown = strings.Replace(markdown, _customMoreTag, "", 1)
		log.Warn().
			Msgf("Manual summary splitting is not supported: %s", page.metadata)
	}

	markdown = ReplaceAbsoluteLinksWithRelative(page.absoluteURL.Host, markdown)
	markdown = replaceCatlistWithShortcode(markdown)
	// Disabled for now, as it does not work well
	if false {
		markdown = highlightCode(markdown)
	} else {
		log.Debug().Msg("Auto-detecting languages of code blocks is disabled for now")
	}
	return &markdown, nil
}

// Markdown converter will automatically pick up "class" attribute fromn "code" tag
// Ref: https://github.com/JohannesKaufmann/html-to-markdown/blob/master/commonmark.go#L334
func improvePreTagsWithCode(htmlContent string) string {
	// Replace all occurrences of "data-enlighter-language" with "language"
	if strings.Contains(htmlContent, "data-enlighter-language") {
		htmlContent = strings.ReplaceAll(htmlContent, `data-enlighter-language="golang"`, `data-enlighter-language="go"`)
		htmlContent = strings.ReplaceAll(htmlContent, `data-enlighter-language="shell"`, `data-enlighter-language="bash"`)
		htmlContent = strings.ReplaceAll(htmlContent, `data-enlighter-language="sh"`, `data-enlighter-language="bash"`)
		htmlContent = strings.ReplaceAll(htmlContent, `data-enlighter-language="lang:`, `data-enlighter-language="`)
		htmlContent = strings.ReplaceAll(htmlContent, `data-enlighter-language="language-`, `data-enlighter-language="`)
		htmlContent = strings.ReplaceAll(htmlContent, `data-enlighter-language="raw"`, "")
		htmlContent = strings.ReplaceAll(htmlContent, `data-enlighter-language="generic"`, "")
		htmlContent = _preTagExtractor1.ReplaceAllString(htmlContent, `<pre><code class="$1">$2</code></pre>`)
		htmlContent = strings.ReplaceAll(htmlContent, `class="EnlighterJSRAW"`, "")
	}
	if strings.Contains(htmlContent, "pre class=") {
		htmlContent = _preTagExtractor2.ReplaceAllString(htmlContent, `<pre><code class="$1">$2</code></pre>`)
	}
	return htmlContent
}

// Mark code blocks with auto-detected language
// Note: https://github.com/alecthomas/chroma is fairly inaccurate in detecting languages
func highlightCode(markdown string) string {
	var _codeBlocExtractor = regexp.MustCompile("\\`\\`\\`(.*)?\\n([.\\s\\S]*)?\\`\\`\\`")
	matches := _codeBlocExtractor.FindAllStringSubmatch(markdown, -1)
	if len(matches) == 0 {
		log.Debug().
			Msg("[highlightCode]No code blocks found")
	}
	for _, match := range matches {
		language := match[1]
		code := match[2]
		// Some WordPress code blocks have language set as "none"!
		if language != "" && language != "none" {
			log.Debug().
				Str("language", language).
				Msg("Code block already has a language")
			continue
		}
		language = getLanguageCode(code)
		if language == "" {
			continue
		}
		code = fmt.Sprintf("```%s\n%s\n```", language, code)
		markdown = strings.Replace(markdown, match[0], code, 1)
	}
	return markdown
}

func getLanguageCode(code string) string {
	possibleLanguages := []string{"Go", "Python", "Java", "C", "Shell", "HTML", "JSON", "YAML"}
	languageCodes := []string{"go", "py", "js", "ts", "java", "c", "sh", "html", "json", "yaml"}

	// enry cannot detect Go language by content!
	if strings.Contains(code, "go.mod") ||
		strings.Contains(code, "go.sum") ||
		strings.Contains(code, "go run") {
		return "go"
	}

	language, onlyOne := enry.GetLanguageByClassifier([]byte(code), possibleLanguages)
	if language == "" {
		log.Warn().
			Str("code", code).
			Msg("No language detected for code block")
		return ""
	}
	if !onlyOne {
		log.Warn().
			Str("code", code).
			Str("language", language).
			Msg("Multiple languages detected for code block")
	}
	log.Debug().
		Str("code", code).
		Str("language", language).
		Msg("Detected language for code block")
	return languageCodes[slices.Index(possibleLanguages, language)]
}

func (page Page) writeContent(w io.Writer) error {
	if _, err := w.Write([]byte(page.markdown)); err != nil {
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
