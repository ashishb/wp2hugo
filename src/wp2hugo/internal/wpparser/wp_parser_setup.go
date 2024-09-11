package wpparser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode"

	ext "github.com/mmcdole/gofeed/extensions"
	"github.com/mmcdole/gofeed/rss"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	errTrashItem         = fmt.Errorf("item is in trash")
	nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)
)

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
}

type WebsiteInfo struct {
	Title       string
	Link        string
	Description string

	PubDate  *time.Time
	Language string

	Categories []CategoryInfo
	Tags       []TagInfo

	// Collecting attachments is mostly useless but we are doing it for completeness
	// Only the ones that are actually used in posts/pages are useful
	Attachments     []AttachmentInfo
	Pages           []PageInfo
	Posts           []PostInfo
	NavigationLinks []NavigationLink
}

type NavigationLink struct {
	// Fallback to Label if title is empty
	Title string
	URL   string
	Type  string
}

type CategoryInfo struct {
	ID       string
	Name     string
	NiceName string
}

type TagInfo struct {
	ID   string
	Name string
	Slug string
}

type PublishStatus string

// See some discussion here https://github.com/ashishb/wp2hugo/issues/26
const (
	PublishStatusAttachment PublishStatus = "attachment"
	PublishStatusDraft      PublishStatus = "draft"
	PublishStatusFuture     PublishStatus = "future"
	PublishStatusInherit    PublishStatus = "inherit"
	PublishStatusPending    PublishStatus = "pending"
	PublishStatusPrivate    PublishStatus = "private"
	PublishStatusPublish    PublishStatus = "publish"
	PublishStatusStatic     PublishStatus = "static"
	PublishStatusTrash      PublishStatus = "trash"
)

type CommonFields struct {
	PostID string

	Author           string
	Title            string
	Link             string     // Note that this is the absolute link for example https://example.com/about
	PublishDate      *time.Time // This can be nil since an item might have never been published
	LastModifiedDate *time.Time
	PublishStatus    PublishStatus // "publish", "draft", "pending" etc. may be make this a custom type
	GUID             *rss.GUID
	PostFormat       *string

	Description string // how to use this?
	Content     string
	Excerpt     string // may be empty

	Categories      []string
	Tags            []string
	Footnotes       []Footnote
	FeaturedImageID *string // Optional WordPress attachment ID of the featured image

	attachmentURL *string
}

func (i CommonFields) Filename() string {
	str1 := strings.ToLower(i.Title)

	// Remove diacritics
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(t, str1)
	if err != nil {
		log.Warn().
			Str("title", i.Title).
			Err(err).
			Msgf("error removing diacritics from title")
	} else {
		str1 = result
	}

	str1 = nonAlphanumericRegex.ReplaceAllString(str1, "-")
	for strings.Contains(str1, "--") {
		str1 = strings.ReplaceAll(str1, "--", "-")
		str1 = strings.ReplaceAll(str1, "-.", ".")
	}
	// Remove leading and trailing "-"
	if len(str1) > 1 {
		str1 = strings.TrimPrefix(str1, "-")
	}
	if len(str1) > 1 {
		str1 = strings.TrimSuffix(str1, "-")
	}
	return str1
}

func (i CommonFields) GetAttachmentURL() *string {
	return i.attachmentURL
}

type PageInfo struct {
	CommonFields
}

type PostInfo struct {
	CommonFields
}

type AttachmentInfo struct {
	CommonFields
}

type Footnote struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

// Parse parses the XML data and returns the WebsiteInfo.
// authors is a list of author names. If it is empty, all authors are considered.
func (p *Parser) Parse(xmlData io.Reader, authors []string) (*WebsiteInfo, error) {
	fp := rss.Parser{}
	feed, err := fp.Parse(InvalidatorCharacterRemover{reader: xmlData})
	if err != nil {
		log.Warn().
			Err(err).
			Msgf("error parsing XML")
		return nil, fmt.Errorf("error parsing XML: %s", err)
	}
	nonEmptyAuthors := make([]string, 0, len(authors))
	for _, a := range authors {
		a = strings.TrimSpace(a)
		if a != "" {
			nonEmptyAuthors = append(nonEmptyAuthors, a)
		}
	}
	return p.getWebsiteInfo(feed, nonEmptyAuthors)
}

func (p *Parser) getWebsiteInfo(feed *rss.Feed, authors []string) (*WebsiteInfo, error) {
	if feed.PubDateParsed == nil {
		log.Warn().Msgf("error parsing published date: %s", feed.PubDateParsed)
	}

	log.Trace().
		Any("WordPress specific keys", keys(feed.Extensions["wp"])).
		Any("Term", feed.Extensions["wp"]["term"]).
		Msg("feed.Custom")

	categories := getCategories(feed.Extensions["wp"]["category"])
	tags := getTags(feed.Extensions["wp"]["tag"])

	attachments := make([]AttachmentInfo, 0)
	pages := make([]PageInfo, 0)
	posts := make([]PostInfo, 0)
	var navigationLinks []NavigationLink

	for _, item := range feed.Items {
		wpPostType := item.Extensions["wp"]["post_type"][0].Value
		switch wpPostType {
		case "attachment":
			if attachment, err := getAttachmentInfo(item); err != nil && !errors.Is(err, errTrashItem) {
				return nil, err
			} else if attachment != nil && hasValidAuthor(authors, attachment.CommonFields) {
				attachments = append(attachments, *attachment)
			}
		case "page":
			if page, err := getPageInfo(item); err != nil && !errors.Is(err, errTrashItem) {
				return nil, err
			} else if page != nil {
				if page.Content == "" && hasValidAuthor(authors, page.CommonFields) {
					log.Warn().
						Str("title", page.Title).
						Msg("Empty content")
				}
				pages = append(pages, *page)
			}
		case "post":
			if post, err := getPostInfo(item); err != nil && !errors.Is(err, errTrashItem) {
				return nil, err
			} else if post != nil && hasValidAuthor(authors, post.CommonFields) {
				if post.Content == "" {
					log.Warn().
						Str("title", post.Title).
						Msg("Empty content")
				}
				posts = append(posts, *post)
			}
		case "wp_navigation":
			var err error
			navigationLinks, err = getNavigationLinks(item.Content)
			if err != nil {
				return nil, fmt.Errorf("error getting navigation links: %w", err)
			}
		case "amp_validated_url", "nav_menu_item", "custom_css", "wp_global_styles":
			// Ignoring these for now
			continue
		default:
			log.Info().
				Str("title", item.Title).
				Str("type", wpPostType).
				Msg("Ignoring item")
		}
	}

	websiteInfo := WebsiteInfo{
		Title:       feed.Title,
		Link:        feed.Link,
		Description: feed.Description,
		PubDate:     feed.PubDateParsed,
		Language:    feed.Language,

		Categories: categories,
		Tags:       tags,

		Attachments:     attachments,
		Pages:           pages,
		Posts:           posts,
		NavigationLinks: navigationLinks,
	}
	log.Info().
		Int("numAttachments", len(websiteInfo.Attachments)).
		Int("numPages", len(websiteInfo.Pages)).
		Int("numPosts", len(websiteInfo.Posts)).
		Int("numCategories", len(categories)).
		Int("numTags", len(tags)).
		Msgf("WebsiteInfo: %s", websiteInfo.Title)
	return &websiteInfo, nil
}

func getAttachmentInfo(item *rss.Item) (*AttachmentInfo, error) {
	fields, err := getCommonFields(item)
	if err != nil {
		return nil, fmt.Errorf("error getting common fields: %w", err)
	}
	attachment := AttachmentInfo{*fields}
	log.Trace().
		Any("attachment", attachment).
		Msg("Attachment")
	return &attachment, nil
}

func getCommonFields(item *rss.Item) (*CommonFields, error) {
	var lastModifiedDate *time.Time
	values := item.Extensions["wp"]["post_modified_gmt"]
	if len(values) > 0 {
		var err error
		lastModifiedDate, err = parseTime(values[0].Value)
		if err != nil {
			log.Warn().
				Str("link", item.Link).
				Str("date", item.Extensions["wp"]["post_modified_gmt"][0].Value).
				Err(err).
				Msg("Error parsing last modified date")
		}
	}

	publishStatus := item.Extensions["wp"]["status"][0].Value
	switch PublishStatus(publishStatus) {
	case PublishStatusAttachment, PublishStatusDraft, PublishStatusFuture, PublishStatusInherit, PublishStatusPending,
		PublishStatusPrivate, PublishStatusPublish, PublishStatusStatic:
		// OK
	case PublishStatusTrash:
		return nil, fmt.Errorf("%w, ignored: %s", errTrashItem, item.Title)
	default:
		log.Fatal().Msgf("Unknown publish status: '%s' for '%s'", publishStatus, item.Title)
	}
	pageCategories := make([]string, 0, len(item.Categories))
	pageTags := make([]string, 0, len(item.Categories))
	var postFormat *string

	for _, category := range item.Categories {
		if isCategory(category) {
			pageCategories = append(pageTags, NormalizeCategoryName(category.Value))
		} else if isTag(category) {
			pageTags = append(pageTags, NormalizeCategoryName(category.Value))
		} else if isPostFormat(category) {
			tmp := NormalizeCategoryName(category.Value)
			postFormat = &tmp
		} else {
			log.Warn().
				Str("link", item.Link).
				Any("categories", item.Categories).
				Msgf("Unknown category: %s", category)
		}
	}
	if len(item.Links) > 1 {
		log.Warn().
			Str("link", item.Link).
			Any("links", item.Links).
			Msg("Multiple links are not handled right now")
	}

	var attachmentURL *string
	tmp1 := item.Extensions["wp"]["attachment_url"]
	if len(tmp1) > 0 {
		attachmentURL = &tmp1[0].Value
		log.Debug().
			Str("attachmentURL", *attachmentURL).
			Msg("Attachment URL")
	}

	pubDate := item.PubDateParsed
	if pubDate == nil && len(item.Extensions["wp"]["post_date"]) > 0 {
		tmp, err := time.Parse("2006-01-02 15:04:05", item.Extensions["wp"]["post_date"][0].Value)
		if err != nil {
			log.Warn().
				Str("link", item.Link).
				Str("date", item.Extensions["wp"]["post_date"][0].Value).
				Msg("Error parsing date")
		} else {
			pubDate = &tmp
		}
	}

	return &CommonFields{
		Author:           getAuthor(item),
		PostID:           item.Extensions["wp"]["post_id"][0].Value,
		Title:            item.Title,
		Link:             item.Link,
		PublishDate:      pubDate,
		GUID:             item.GUID,
		LastModifiedDate: lastModifiedDate,
		PublishStatus:    PublishStatus(publishStatus),
		PostFormat:       postFormat,
		Excerpt:          item.Extensions["excerpt"]["encoded"][0].Value,

		Description:     item.Description,
		Content:         item.Content,
		Categories:      pageCategories,
		Tags:            pageTags,
		Footnotes:       getFootnotes(item),
		FeaturedImageID: getThumbnailID(item),

		attachmentURL: attachmentURL,
	}, nil
}

func hasValidAuthor(authors []string, fields CommonFields) bool {
	if len(authors) == 0 {
		return true
	}
	for _, a := range authors {
		if a == fields.Author {
			return true
		}
	}
	log.Warn().
		Str("author", fields.Author).
		Str("authors", strings.Join(authors, ",")).
		Str("title", fields.Title).
		Str("link", fields.Link).
		Msg("Author not in the list of authors to process")
	return false
}

func getAuthor(item *rss.Item) string {
	author := item.Author
	if len(author) > 0 {
		return author
	}
	if item.Extensions["dc"] != nil && item.Extensions["dc"]["creator"] != nil {
		return item.Extensions["dc"]["creator"][0].Value
	}
	return ""
}

func getNavigationLinks(content string) ([]NavigationLink, error) {
	// Extract all HTML comments
	var htmlCommentExtractor = regexp.MustCompile(`<!--(.*?)-->`)
	comments := htmlCommentExtractor.FindAllString(content, -1)
	log.Debug().
		Int("navigationLinks", len(comments)).
		Msg("getNavigationLinks")
	results := make([]NavigationLink, 0, len(comments))
	for _, comment := range comments {
		log.Trace().Msgf("comment: %s", comment)
		var navigationLinkExtractor = regexp.MustCompile(`{.*}`)
		match := navigationLinkExtractor.FindString(comment)
		if match == "" {
			continue
		}
		link, err := getNavigationLink(match)
		if err != nil {
			return nil, fmt.Errorf("error getting navigation link: %w", err)
		}
		log.Debug().
			Any("link", link).
			Msg("Navigation link")
		results = append(results, *link)
	}
	return results, nil
}

/*
*
  - Example:
    {
    "className": " menu-item menu-item-type-taxonomy menu-item-object-category",
    "description": "",
    "id": "113",
    "kind": "taxonomy",
    "label": "Tech thoughts",
    "opensInNewTab": false,
    "rel": null,
    "title": "",
    "type": "category",
    "url": "https://ashishb.net/category/tech-thoughts/"
    }
*/
func getNavigationLink(match string) (*NavigationLink, error) {
	type _NavigationLink struct {
		// Note that ID is string in ashishb.net export but is int in some other exports
		// https://github.com/ashishb/wp2hugo/issues/9
		Label string `json:"label"`
		Title string `json:"title"`
		Type  string `json:"type"`
		URL   string `json:"url"`
	}
	var navLink _NavigationLink
	if err := json.Unmarshal([]byte(match), &navLink); err != nil {
		return nil, fmt.Errorf("error unmarshalling navigation link: %w", err)
	}
	title := navLink.Title
	if title == "" {
		title = navLink.Label
	}
	return &NavigationLink{
		Title: title,
		URL:   navLink.URL,
		Type:  navLink.Type,
	}, nil
}

func isCategory(category *rss.Category) bool {
	return category.Domain == "category"
}

func isPostFormat(category *rss.Category) bool {
	return category.Domain == "post_format"
}

func isTag(tag *rss.Category) bool {
	return tag.Domain == "post_tag"
}

// NormalizeCategoryName removes space from the category name and converts it to lowercase
func NormalizeCategoryName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}

func getPageInfo(item *rss.Item) (*PageInfo, error) {
	fields, err := getCommonFields(item)
	if err != nil {
		return nil, fmt.Errorf("error getting common fields: %w", err)
	}
	page := PageInfo{*fields}
	log.Trace().
		Any("page", page).
		Msg("Page")
	return &page, nil
}

func getPostInfo(item *rss.Item) (*PostInfo, error) {
	fields, err := getCommonFields(item)
	if err != nil {
		return nil, fmt.Errorf("error getting common fields: %w", err)
	}
	post := PostInfo{*fields}
	log.Trace().
		Any("post", post).
		Msg("Post")
	return &post, nil
}

func getCategories(inputs []ext.Extension) []CategoryInfo {
	categories := make([]CategoryInfo, 0, len(inputs))
	for _, input := range inputs {
		category := CategoryInfo{
			// ID is usually int but for safety let's assume string
			ID:       input.Children["term_id"][0].Value,
			Name:     NormalizeCategoryName(input.Children["cat_name"][0].Value),
			NiceName: input.Children["category_nicename"][0].Value,
			// We are ignoring "category_parent" for now as I have never used it
		}
		log.Trace().Msgf("category: %+v", category)
		categories = append(categories, category)
	}
	return categories
}

func getTags(inputs []ext.Extension) []TagInfo {
	categories := make([]TagInfo, 0, len(inputs))
	for _, input := range inputs {
		var tagName string
		if len(input.Children["tag_name"]) == 0 {
			// Fallback
			tagName = input.Children["tag_slug"][0].Value
			log.Warn().
				Any("input", input).
				Msg("tag_name is missing")
		} else {
			tagName = input.Children["tag_name"][0].Value
		}
		tag := TagInfo{
			// ID is usually int but for safety let's assume string
			ID:   input.Children["term_id"][0].Value,
			Name: NormalizeCategoryName(tagName),
			Slug: input.Children["tag_slug"][0].Value,
		}
		log.Trace().Msgf("tag: %+v", tag)
		categories = append(categories, tag)
	}
	return categories
}

func getFootnotes(item *rss.Item) []Footnote {
	if len(item.Extensions["wp"]["postmeta"]) == 0 {
		return nil
	}

	footnotes := make([]Footnote, 0)
	// Footnotes
	for _, meta := range item.Extensions["wp"]["postmeta"] {
		if len(meta.Children["meta_key"]) == 0 {
			continue
		}
		if len(meta.Children["meta_value"]) == 0 {
			continue
		}
		if meta.Children["meta_key"][0].Value != "footnotes" {
			continue
		}
		if len(meta.Children["meta_value"][0].Value) == 0 {
			log.Warn().
				Str("link", item.Link).
				Msg("ignoring empty footnote")
			continue
		}
		footnoteJSON := meta.Children["meta_value"][0].Value
		footnoteArr := make([]Footnote, 0)
		if err := json.Unmarshal([]byte(footnoteJSON), &footnoteArr); err != nil {
			log.Warn().
				Str("link", item.Link).
				Str("footnoteJSON", footnoteJSON).
				Err(err).
				Msg("Error unmarshalling footnotes")
		} else {
			log.Debug().
				Any("footnotes", footnotes).
				Msg("Footnotes")
			footnotes = append(footnotes, footnoteArr...)
		}
	}
	if len(footnotes) == 0 {
		return nil
	}
	log.Debug().
		Int("numFootnotes", len(footnotes)).
		Str("link", item.Link).
		Msg("Footnotes found")
	return footnotes
}

func getThumbnailID(item *rss.Item) *string {
	if len(item.Extensions["wp"]["postmeta"]) == 0 {
		return nil
	}

	for _, meta := range item.Extensions["wp"]["postmeta"] {
		if len(meta.Children["meta_key"]) == 0 {
			continue
		}
		if len(meta.Children["meta_value"]) == 0 {
			continue
		}
		if meta.Children["meta_key"][0].Value != "_thumbnail_id" {
			continue
		}
		thumbnailID := meta.Children["meta_value"][0].Value
		log.Debug().
			Str("thumbnailID", thumbnailID).
			Msg("Thumbnail ID")
		return &thumbnailID
	}
	return nil
}

func parseTime(utcTime string) (*time.Time, error) {
	t, err := time.Parse("2006-01-02 15:04:05", utcTime)
	if err != nil {
		return nil, fmt.Errorf("error parsing time: %w", err)
	}
	return &t, nil
}

// keys returns the keys of the map m.
// The keys will be an indeterminate order.
func keys[M ~map[K]V, K comparable, V any](m M) []K {
	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	return r
}
