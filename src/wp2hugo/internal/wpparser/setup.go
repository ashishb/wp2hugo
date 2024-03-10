package wpparser

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mmcdole/gofeed/extensions"
	"github.com/mmcdole/gofeed/rss"
	"github.com/rs/zerolog/log"
	"io"
	"regexp"
	"strings"
	"time"
)

var (
	errTrashItem         = fmt.Errorf("item is in trash")
	nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
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

const (
	PublishStatusPublish PublishStatus = "publish"
	PublishStatusDraft   PublishStatus = "draft"
	PublishStatusPending PublishStatus = "pending"
	PublishStatusInherit PublishStatus = "inherit"
	PublishStatusFuture  PublishStatus = "future"
)

type CommonFields struct {
	PostID string

	Title            string
	Link             string     // Note that this is the absolute link for example https://example.com/about
	PublishDate      *time.Time // This can be nil since an item might have never been published
	LastModifiedDate *time.Time
	PublishStatus    PublishStatus // "publish", "draft", "pending" etc. may be make this a custom type

	Description string // how to use this?
	Content     string
	Excerpt     string // may be empty

	Categories []string
	Tags       []string

	// TODO: may be add author
}

func (i CommonFields) Filename() string {
	str1 := strings.ToLower(i.Title)
	str1 = nonAlphanumericRegex.ReplaceAllString(str1, "_")
	for strings.Contains(str1, "__") {
		str1 = strings.ReplaceAll(str1, "__", "_")
	}
	return str1
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

func (p *Parser) Parse(xmlData io.Reader) (*WebsiteInfo, error) {
	fp := rss.Parser{}
	feed, err := fp.Parse(InvalidatorCharacterRemover{reader: xmlData})
	if err != nil {
		return nil, fmt.Errorf("error parsing XML: %s", err)
	}
	return p.getWebsiteInfo(feed)
}

func (p *Parser) getWebsiteInfo(feed *rss.Feed) (*WebsiteInfo, error) {
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
			} else if attachment != nil {
				attachments = append(attachments, *attachment)
			}
		case "page":
			if page, err := getPageInfo(item); err != nil && !errors.Is(err, errTrashItem) {
				return nil, err
			} else if page != nil {
				if page.Content == "" {
					log.Warn().
						Str("title", page.Title).
						Msg("Empty content")
				}
				pages = append(pages, *page)
			}
		case "post":
			if post, err := getPostInfo(item); err != nil && !errors.Is(err, errTrashItem) {
				return nil, err
			} else if post != nil {
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
	lastModifiedDate, err := parseTime(item.Extensions["wp"]["post_modified_gmt"][0].Value)
	if err != nil {
		return nil, fmt.Errorf("error parsing last modified date: %w", err)
	}

	publishStatus := item.Extensions["wp"]["status"][0].Value
	switch publishStatus {
	case "publish", "draft", "pending", "inherit", "future":
		// OK
	case "trash":
		return nil, fmt.Errorf("%w, ignored: %s", errTrashItem, item.Title)
	default:
		log.Fatal().Msgf("Unknown publish status: %s for %s", publishStatus, item.Title)
	}
	pageCategories := make([]string, 0, len(item.Categories))
	pageTags := make([]string, 0, len(item.Categories))

	for _, category := range item.Categories {
		if isCategory(category) {
			pageCategories = append(pageTags, category.Value)
		} else if isTag(category) {
			pageTags = append(pageTags, category.Value)
		} else {
			log.Fatal().
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

	return &CommonFields{
		PostID:           item.Extensions["wp"]["post_id"][0].Value,
		Title:            item.Title,
		Link:             item.Link,
		PublishDate:      item.PubDateParsed,
		LastModifiedDate: lastModifiedDate,
		PublishStatus:    PublishStatus(publishStatus),
		Excerpt:          item.Extensions["excerpt"]["encoded"][0].Value,

		Description: item.Description,
		Content:     item.Content,
		Categories:  pageCategories,
		Tags:        pageTags,
	}, nil
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
		ClassName     string `json:"className"`
		Description   string `json:"description"`
		ID            string `json:"id"`
		Kind          string `json:"kind"`
		Label         string `json:"label"`
		OpensInNewTab bool   `json:"opensInNewTab"`
		Rel           string `json:"rel"`
		Title         string `json:"title"`
		Type          string `json:"type"`
		URL           string `json:"url"`
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

func isTag(tag *rss.Category) bool {
	return tag.Domain == "post_tag"
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
			Name:     input.Children["cat_name"][0].Value,
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
		tag := TagInfo{
			// ID is usually int but for safety let's assume string
			ID:   input.Children["term_id"][0].Value,
			Name: input.Children["tag_name"][0].Value,
			Slug: input.Children["tag_slug"][0].Value,
		}
		log.Trace().Msgf("tag: %+v", tag)
		categories = append(categories, tag)
	}
	return categories
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
