package hugopage

import (
	"fmt"
	"io"
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/go-enry/go-enry/v2"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
	"github.com/leeqvip/gophp"
	"github.com/mmcdole/gofeed/rss"
	"github.com/rs/zerolog/log"
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
	attachments []wpparser.AttachmentInfo

	metadata map[string]any
	markdown string
}

const _WordPressMoreTag = "<!--more-->"

// In the next step, we will replace this as well
const (
	_customMoreTag   = "{{< more >}}"
	_wordPressTocTag = "[toc]"
)

var (
	// E.g. <pre class="EnlighterJSRAW" data-enlighter-language="golang">
	_preTagExtractor1 = regexp.MustCompile(`<pre class="EnlighterJSRAW" data-enlighter-language="([^"]+?)".*?>([\s\S]*?)</pre>`)
	// E.g. <pre class="lang:bash" nums="false">
	_preTagExtractor2 = regexp.MustCompile(`<pre class=".*?lang:([^" ]+).*?>([\s\S]*?)</pre>`)

	_hugoShortCodeMatcher = regexp.MustCompile(`{{<.*?>}}`)

	// Ref: https://github.com/markdownlint/markdownlint/blob/main/docs/RULES.md#md012---multiple-consecutive-blank-lines
	// Replace multiple consecutive newlines with just two newlines
	_moreThanTwoNewlines = regexp.MustCompile(`\n{3,}`)

	// Catch \nspace\n
	_spaceSurroundedByNewlines = regexp.MustCompile(`\n[ \t]+\n`)
)

// Extracts "src" from Hugo figure shortcode
// {{< figure align=aligncenter width=905 src="/wp-content/uploads/2023/01/Stollemeyer-castle-1024x768.jpg" alt="" >}}
var _hugoFigureLinks = regexp.MustCompile(`{{< figure.*?src="([^\"]+?)".*? >}}`)

// Extracts "src" from Hugo audio shortcode
// {{< audio src="/wp-content/uploads/2023/01/session.mp3" alt="" >}}
var _hugoAudioLinks = regexp.MustCompile(`{{< audio.*?src="([^\"]+?)".*? >}}`)

// {{< parallaxblur src="/wp-content/uploads/2018/12/bora%5Fbora%5F5%5Fresized.jpg" >}}
var _hugoParallaxBlurLinks = regexp.MustCompile(`{{< parallaxblur.*?src="([^\"]+?)".*? >}}`)

func NewPage(provider ImageURLProvider, pageURL url.URL, author string, title string, publishDate *time.Time,
	isDraft bool, categories []string, tags []string, attachments []wpparser.AttachmentInfo,
	footnotes []wpparser.Footnote,
	htmlContent string, guid *rss.GUID, featuredImageID *string, postFormat *string,
	customMetaData []wpparser.CustomMetaDatum, taxinomies []wpparser.TaxonomyInfo,
	postID string, parentPostID *string,
) (*Page, error) {
	metadata, err := getMetadata(provider, pageURL, author, title, publishDate, isDraft, categories, tags, guid,
		featuredImageID, postFormat, customMetaData, taxinomies, postID, parentPostID)
	if err != nil {
		return nil, err
	}
	page := Page{
		absoluteURL: pageURL,
		metadata:    metadata,
		attachments: attachments,
	}
	// htmlContent is the HTML content of the page that will be
	// transformed to Markdown
	markdown, err := page.getMarkdown(provider, htmlContent, footnotes)
	if err != nil {
		return nil, err
	}
	page.markdown = *markdown
	return &page, nil
}

func (page *Page) Markdown() string {
	return page.markdown
}

func (page *Page) Replace(replacementMap map[string]string) {
	for old, new := range replacementMap {
		page.markdown = strings.ReplaceAll(page.markdown, old, new)
	}
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

func (page *Page) WPMediaLinks() []string {
	arr1 := getImageLinks([]byte(page.markdown))
	arr2 := getMarkdownLinks(_hugoFigureLinks, page.markdown)
	arr3 := getMarkdownLinks(_hugoParallaxBlurLinks, page.markdown)
	arr4 := getMarkdownLinks(_hugoAudioLinks, page.markdown)
	arr5 := getPDFLinks([]byte(page.markdown))
	coverImageURL := page.getCoverImageURL()
	result := append(append(append(append(arr1, arr2...), arr3...), arr4...), arr5...)
	if coverImageURL != nil {
		result = append(result, *coverImageURL)
	}
	return result
}

func getImageLinks(content []byte) []string {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := markdown.Parse(content, p)

	var links []string
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if img, ok := node.(*ast.Image); ok && entering {
			links = append(links, string(img.Destination))
		}
		return ast.GoToNext
	})

	log.Debug().
		Int("count", len(links)).
		Msg("Image links")
	return links
}

func getPDFLinks(content []byte) []string {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := markdown.Parse(content, p)

	var links []string
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if link, ok := node.(*ast.Link); ok && entering {
			destination := string(link.Destination)
			if strings.HasSuffix(strings.ToLower(destination), ".pdf") {
				links = append(links, destination)
			}
		}
		return ast.GoToNext
	})

	log.Debug().
		Int("count", len(links)).
		Msg("PDF links")
	return links
}

func getMarkdownLinks(regex *regexp.Regexp, markdown string) []string {
	matches := regex.FindAllStringSubmatch(markdown, -1)
	links := make([]string, 0, len(matches))
	for _, match := range matches {
		links = append(links, match[1])
	}
	return links
}

func unserialiazePHParray(array string) any {
	/* Ex:
	a:2:{s:10:"taxonomies";s:32:"f166db6f0df2a3df4c2715a8bcc30eec";s:15:"postmeta_fields";s:32:"0edff5c6e53a54394f90f7b5a8fc1e76";}
	*/
	phpArray, err := gophp.Unserialize([]byte(array))
	if err != nil {
		log.Error().
			Err(err).
			Str("array", array).
			Msg("Failed to decode PHP serialized array")
		return nil
	} else {
		return phpArray
	}
}

func getMetadata(provider ImageURLProvider, pageURL url.URL, author string, title string, publishDate *time.Time,
	isDraft bool, categories []string, tags []string, guid *rss.GUID, featuredImageID *string,
	postFormat *string, customMetaData []wpparser.CustomMetaDatum, taxinomies []wpparser.TaxonomyInfo,
	postID string, parentPostID *string,
) (map[string]any, error) {
	metadata := make(map[string]any)
	metadata["url"] = pageURL.Path // Relative URL
	metadata["author"] = author
	metadata["title"] = title
	metadata["post_id"] = postID
	metadata["parent_post_id"] = parentPostID
	if publishDate != nil {
		metadata["date"] = publishDate.Format(_hugoDateFormat)
	}
	if isDraft {
		metadata["draft"] = "true"
	}
	if len(categories) > 0 {
		sort.Strings(categories)
		metadata[CategoryName] = slices.Compact(categories)
	}
	if len(tags) > 0 {
		sort.Strings(tags)
		metadata[TagName] = slices.Compact(tags)
	}

	for _, taxinomy := range taxinomies {
		if existing, ok := metadata[taxinomy.Taxonomy]; ok {
			switch v := existing.(type) {
			case []string:
				metadata[taxinomy.Taxonomy] = append(v, taxinomy.Name)
			default:
				metadata[taxinomy.Taxonomy] = []string{taxinomy.Name}
			}
		} else {
			metadata[taxinomy.Taxonomy] = []string{taxinomy.Name}
		}
	}

	for _, metadatum := range customMetaData {
		if strings.HasPrefix(metadatum.Value, "a:") {
			phpArray := unserialiazePHParray(metadatum.Value)
			if phpArray != nil {
				// Decoded array is a nested dictionnary
				metadata[metadatum.Key] = phpArray
			} else {
				// Fallback to ugly serialized array
				metadata[metadatum.Key] = metadatum.Value
			}
		} else {
			// Simple string
			metadata[metadatum.Key] = metadatum.Value
		}
		// Note: if any step of the PHP array reading/decoding failed,
		// now we got the original serialized array
	}

	if guid != nil {
		metadata["guid"] = guid.Value
	}

	if featuredImageID != nil {
		if imageInfo, err := provider.GetImageInfo(*featuredImageID); err != nil {
			log.Warn().
				Err(err).
				Str("imageID", *featuredImageID).
				Msg("Image URL not found")
		} else {
			coverInfo := make(map[string]string)
			imageURL, err := url.Parse(imageInfo.ImageURL)
			if err != nil {
				return nil, fmt.Errorf("error parsing image URL '%s': %w", imageInfo.ImageURL, err)
			}
			if imageURL.Host == pageURL.Host {
				// If the image URL is on the same host as the page, we can use a relative URL
				coverInfo["image"] = imageURL.Path
			} else {
				coverInfo["image"] = imageInfo.ImageURL
			}
			coverInfo["alt"] = imageInfo.Title
			metadata["cover"] = coverInfo
		}
	}
	if postFormat != nil {
		metadata["type"] = *postFormat
	}
	return metadata, nil
}

func (page *Page) getCoverImageURL() *string {
	if page.metadata == nil {
		return nil
	}
	cover, ok := page.metadata["cover"]
	if !ok {
		return nil
	}
	coverInfo, ok := cover.(map[string]string)
	if !ok {
		return nil
	}
	url1, ok := coverInfo["image"]
	if !ok {
		return nil
	}
	return &url1
}

func (page *Page) writeMetadata(w io.Writer) error {
	combinedMetadata, err := utils.GetYAML(page.metadata)
	if err != nil {
		return fmt.Errorf("error marshalling metadata: %w", err)
	}
	combinedMetadataStr := fmt.Sprintf("---\n%s\n---\n", string(combinedMetadata))
	if _, err := w.Write([]byte(combinedMetadataStr)); err != nil {
		return fmt.Errorf("error writing to page file: %w", err)
	}
	return nil
}

func (page *Page) getMarkdown(provider ImageURLProvider, htmlContent string, footnotes []wpparser.Footnote) (*string, error) {
	if htmlContent == "" {
		log.Error().
			Any("page", page.metadata).
			Msg("Empty HTML body for page")
		msg := ""
		return &msg, nil
	}
	attachmentIDs := make([]string, 0, len(page.attachments))
	for _, attachment := range page.attachments {
		attachmentIDs = append(attachmentIDs, attachment.PostID)
	}

	converter := getMarkdownConverter()
	htmlContent = improvePreTagsWithCode(htmlContent)
	htmlContent = replaceCaptionWithFigure(htmlContent)
	htmlContent = replaceImageBlockWithFigure(htmlContent)
	htmlContent = replaceAudioShortCode(htmlContent)
	htmlContent = replaceGutembergGalleryWithFigure(htmlContent)
	htmlContent = replaceGalleryWithFigure(provider, attachmentIDs, htmlContent)
	htmlContent = replaceAWBWithParallaxBlur(provider, htmlContent)
	htmlContent = strings.Replace(htmlContent, _WordPressMoreTag, _customMoreTag, 1)

	// We convert consecutive <br> to a custom tag
	// then we convert <br> to "  \n" and then we convert the custom tag to "\n\n"
	// It is convoluted but it works.
	htmlContent = convertConsecutiveBRToCustomTag(htmlContent)

	// This handling is specific to paperMod theme
	// Ref: https://adityatelange.github.io/hugo-PaperMod/posts/papermod/papermod-features/#show-table-of-contents-toc-on-blog-post
	if strings.Contains(htmlContent, _wordPressTocTag) {
		htmlContent = strings.Replace(htmlContent, _wordPressTocTag, "", 1)
		page.metadata["ShowToc"] = true
		page.metadata["TocOpen"] = true
	}
	markdown, err := converter.ConvertString(htmlContent)
	log.Debug().
		Str("htmlContent", htmlContent).
		Str("markdown", markdown).
		Msg("Markdown conversion")

	if err != nil {
		return nil, fmt.Errorf("error converting HTML to Markdown: %w", err)
	}
	if len(strings.TrimSpace(markdown)) == 0 {
		// The page contains no markdown. Warn the user, but keep going.
		log.Warn().
			Str("page", page.absoluteURL.String()).
			Msg("empty markdown")
	}
	if strings.Contains(markdown, _customMoreTag) {
		// Ref: https://gohugo.io/content-management/summaries/#manual-summary-splitting
		summary := strings.Split(markdown, _customMoreTag)[0]
		markdown = strings.Replace(markdown, _customMoreTag, "", 1)
		// Remove short codes from summary
		// Ref: https://github.com/ashishb/wp2hugo/issues/13
		page.metadata["summary"] = strings.TrimSpace(removeAllHugoShortcodes(summary))
		log.Warn().
			Msgf("Manual summary splitting is not supported: %s", page.metadata)
	}

	markdown = strings.ReplaceAll(markdown, _doubleSpaceWithNewline, "  \n")
	markdown = ReplaceAbsoluteLinksWithRelative(page.absoluteURL.Host, markdown)
	markdown = replaceCatlistWithShortcode(markdown)
	// Disabled for now, as it does not work well
	if false {
		markdown = highlightCode(markdown)
	} else {
		log.Debug().Msg("Auto-detecting languages of code blocks is disabled for now")
	}

	// Replace footnote links with actual Hugo-style footnotes
	// Ref: https://geekthis.net/post/hugo-footnotes-and-citations
	footnoteStrs := make([]string, 0, len(footnotes))
	if len(footnotes) > 0 {
		// [^1]: And that's the footnote.
		for i, footnote := range footnotes {
			tmp := fmt.Sprintf("[^%d]: %s", i+1, footnote.Content)
			footnoteStrs = append(footnoteStrs, tmp)
			regex1 := regexp.MustCompile(fmt.Sprintf(`\[\S+\]\(#%s\)`, footnote.ID))
			markdown = regex1.ReplaceAllString(markdown, fmt.Sprintf(`[^%d]`, i+1))
		}
		markdown = fmt.Sprintf("%s\n\n%s", markdown, strings.Join(footnoteStrs, "\n\n"))
	}

	markdown = replaceOrderedListNumbers(markdown)
	markdown = replaceConsecutiveNewlines(markdown)
	markdown = replacePlaintextYoutubeURL(markdown)
	markdown = removeTrailingSpaces(markdown)

	return &markdown, nil
}

func removeAllHugoShortcodes(summary string) string {
	// Ref: https://gohugo.io/content-management/shortcodes/#remove-shortcodes
	return _hugoShortCodeMatcher.ReplaceAllString(summary, " ")
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
	_codeBlocExtractor := regexp.MustCompile("\\`\\`\\`(.*)?\\n([.\\s\\S]*)?\\`\\`\\`")
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

func replaceOrderedListNumbers(markdown string) string {
	// Ref: https://github.com/markdownlint/markdownlint/blob/main/docs/RULES.md#md029---ordered-list-item-prefix
	// Find all the ordered list items starting with optional whitespaces followed by \d. and replace with 1.
	reg1 := regexp.MustCompile(`(?m)^(\s*)(\d+)\.(\s)`)
	return reg1.ReplaceAllString(markdown, `${1}1.$3`)
}

func replaceConsecutiveNewlines(markdown string) string {
	return _moreThanTwoNewlines.ReplaceAllString(markdown, "\n\n")
}

func removeTrailingSpaces(markdown string) string {
	return _spaceSurroundedByNewlines.ReplaceAllString(markdown, "\n\n")
}

func (page Page) writeContent(w io.Writer) error {
	if _, err := w.Write([]byte(page.markdown)); err != nil {
		return fmt.Errorf("error writing to page file: %w", err)
	}

	if !strings.HasSuffix(page.markdown, "\n") {
		// Add a newline at the end of the file
		if _, err := w.Write([]byte("\n")); err != nil {
			return fmt.Errorf("error writing newline to page file: %w", err)
		}
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
