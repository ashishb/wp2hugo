package wpparser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	ext "github.com/mmcdole/gofeed/extensions"
	"github.com/mmcdole/gofeed/rss"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const _filenameSizeLimit = 200

var (
	errTrashItem = errors.New("item is in trash")
	// \p{L} matches any letter from any language while \w matches only ASCII letters
	nonAlphanumericRegex = regexp.MustCompile(`[^\p{L}]+`)
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
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
	PostType         *string // Custom post types, typically FAQ, portfolio, etc.

	// 1. Only attachments seem to have this
	// 2. "0" seems to be reserved for no parent, we replace that with nil
	PostParentID *string // ID of the parent post, if any

	Description string // how to use this?
	Content     string
	Excerpt     string // may be empty

	Categories      []string
	Tags            []string
	Taxonomies      []TaxonomyInfo
	CustomMetaData  []CustomMetaDatum
	Footnotes       []Footnote
	FeaturedImageID *string // Optional WordPress attachment ID of the featured image

	attachmentURL *string

	Comments []CommentInfo
}

func titleToFilename(title string) string {
	str1 := strings.ToLower(title)

	// Remove diacritics
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(t, str1)
	if err != nil {
		log.Warn().
			Str("title", title).
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

	if len(str1) > _filenameSizeLimit {
		log.Warn().
			Str("title", title).
			Msgf("Filename is too long, truncating to %d characters", _filenameSizeLimit)
		str1 = str1[:_filenameSizeLimit]
	}

	return str1
}

func findSlugAndParams(parts []string) (string, string) {
	file := ""
	params := ""

	// The last chunk is either our page slug or URL parameters
	for i := len(parts) - 1; i > 2; i-- {
		part := parts[i]
		if !strings.HasPrefix(part, "?") && part != "" {
			file = part
			if i+1 < len(parts) && strings.HasPrefix(parts[i+1], "?") {
				params = parts[i+1]
			}
			return file, params
		}
	}

	return file, params
}

type FileInfo struct {
	filename string
	language *string
}

func (f FileInfo) FileNameWithLanguage() string {
	if f.language == nil {
		return f.filename
	}

	return fmt.Sprintf("%s.%s", f.filename, *f.language)
}

func (f FileInfo) FileNameNoLanguage() string {
	return f.filename
}

func (f FileInfo) Language() *string {
	return f.language
}

func (i CommonFields) GetFileInfo() FileInfo {
	// Split canonical link path on /
	parts := strings.Split(strings.TrimRight(i.Link, "/"), "/")
	file, params := findSlugAndParams(parts)

	// WooCommerce products have ugly links like https://website.com/?post_type=product&p=666
	// Nothing meaningful there, but their GUID uses pretty links. Retry then.
	if len(file) == 0 {
		parts = strings.Split(strings.TrimRight(i.GUID.Value, "/"), "/")
		file, params = findSlugAndParams(parts)
	}

	// Remove leading and trailing "-"
	if len(file) > 1 {
		file = strings.TrimPrefix(file, "-")
	}
	if len(file) > 1 {
		file = strings.TrimSuffix(file, "-")
	}
	if len(file) == 0 {
		file = titleToFilename((i.Title))
	}

	// Append language suffix if found in link
	langRegex := regexp.MustCompile(`(?:\?|&)lang=([^&$]+)`)
	langMatch := langRegex.FindStringSubmatch(params)
	var lang *string
	if len(langMatch) > 1 {
		lang = lo.ToPtr(langMatch[1])
	}

	return FileInfo{
		filename: file,
		language: lang,
	}
}

func (i CommonFields) GetAttachmentURL() *string {
	return i.attachmentURL
}

type PageInfo struct {
	CommonFields
}

type CustomPostInfo struct {
	CommonFields
}

type PostInfo struct {
	CommonFields
}

type AttachmentInfo struct {
	CommonFields
}

type CommentInfo struct {
	ID          string     `yaml:"id"`
	AuthorName  string     `yaml:"author_name"`
	AuthorEmail string     `yaml:"author_email"`
	AuthorURL   string     `yaml:"author_url"`
	PublishDate *time.Time `yaml:"published"`
	ParentID    string     `yaml:"parent_id"`
	Content     string     `yaml:"content"`
	PostLink    string     `yaml:"post_url"`
	PostID      string     `yaml:"post_id"`
}

type Footnote struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

type CustomMetaDatum struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Parse parses the XML data and returns the WebsiteInfo.
// authors is a list of author names. If it is empty, all authors are considered.
func (p *Parser) Parse(xmlData io.Reader, authors []string, customPostTypes []string) (*WebsiteInfo, error) {
	fp := rss.Parser{}
	feed, err := fp.Parse(InvalidatorCharacterRemover{reader: xmlData})
	if err != nil {
		log.Warn().
			Err(err).
			Msgf("error parsing XML")
		return nil, fmt.Errorf("error parsing XML: %w", err)
	}
	nonEmptyAuthors := make([]string, 0, len(authors))
	for _, a := range authors {
		a = strings.TrimSpace(a)
		if a != "" {
			nonEmptyAuthors = append(nonEmptyAuthors, a)
		}
	}
	return p.getWebsiteInfo(feed, nonEmptyAuthors, customPostTypes)
}

func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func (p *Parser) getWebsiteInfo(feed *rss.Feed, authors []string, customPostTypes []string) (*WebsiteInfo, error) {
	if feed.PubDateParsed == nil {
		log.Warn().Msgf("error parsing published date: %s", feed.PubDateParsed)
	}

	log.Trace().
		Any("WordPress specific keys", keys(feed.Extensions["wp"])).
		Any("Term", feed.Extensions["wp"]["term"]).
		Msg("feed.Custom")

	categories := getCategories(feed.Extensions["wp"]["category"])
	tags := getTags(feed.Extensions["wp"]["tag"])
	taxonomies := getTaxonomies(feed.Extensions["wp"]["term"])

	attachments := make([]AttachmentInfo, 0)
	pages := make([]PageInfo, 0)
	posts := make([]PostInfo, 0)
	customPosts := make([]CustomPostInfo, 0)
	var navigationLinks []NavigationLink

	for _, item := range feed.Items {
		wpPostType := item.Extensions["wp"]["post_type"][0].Value
		switch wpPostType {
		case "attachment":
			if attachment, err := getAttachmentInfo(item, taxonomies); err != nil && !errors.Is(err, errTrashItem) {
				return nil, err
			} else if attachment != nil && hasValidAuthor(authors, attachment.CommonFields) {
				attachments = append(attachments, *attachment)
				log.Debug().
					Str("postID", attachment.PostID).
					Str("postType", wpPostType).
					Msg("processing attachment")
			}
		case "page":
			if page, err := getPageInfo(item, taxonomies); err != nil && !errors.Is(err, errTrashItem) {
				return nil, err
			} else if page != nil {
				if page.Content == "" && hasValidAuthor(authors, page.CommonFields) {
					log.Warn().
						Str("title", page.Title).
						Msg("Empty content")
				}
				pages = append(pages, *page)
				log.Debug().
					Str("postID", page.PostID).
					Str("postType", wpPostType).
					Msg("processing page")
			}
		case "post":
			if post, err := getPostInfo(item, taxonomies); err != nil && !errors.Is(err, errTrashItem) {
				return nil, err
			} else if post != nil && hasValidAuthor(authors, post.CommonFields) {
				if post.Content == "" {
					log.Warn().
						Str("title", post.Title).
						Msg("Empty content")
				}
				log.Debug().
					Str("postID", post.PostID).
					Str("postType", wpPostType).
					Msg("processing Post")
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
			if contains(customPostTypes, wpPostType) {
				if customPost, err := getCustomPostInfo(item, taxonomies); err != nil && !errors.Is(err, errTrashItem) {
					return nil, err
				} else if customPost != nil {
					if customPost.Content == "" {
						log.Warn().
							Str("title", customPost.Title).
							Msg("Empty content")
					}
					customPosts = append(customPosts, *customPost)
					log.Debug().
						Str("postID", customPost.PostID).
						Str("postType", wpPostType).
						Msg("processing post")
				}
			} else {
				log.Info().
					Str("title", item.Title).
					Str("type", wpPostType).
					Msg("Ignoring item due to unknown type")
			}
		}
	}

	linkURL, err := url.Parse(feed.Link)
	if err != nil {
		return nil, fmt.Errorf("error parsing feed link: %w", err)
	}

	websiteInfo := WebsiteInfo{
		title:       feed.Title,
		link:        linkURL,
		Description: feed.Description,
		pubDate:     feed.PubDateParsed,
		language:    feed.Language,

		categories: categories,
		tags:       tags,
		taxonomies: taxonomies,

		attachments:     attachments,
		pages:           pages,
		posts:           posts,
		customPosts:     customPosts,
		navigationLinks: navigationLinks,

		customPostTypes: customPostTypes,

		postIDToAttachmentCache: getPostIDToAttachmentsMap(attachments),
	}
	log.Info().
		Int("numAttachments", len(websiteInfo.attachments)).
		Int("numPages", len(websiteInfo.pages)).
		Int("numPosts", len(websiteInfo.posts)).
		Int("numCustomPosts", len(websiteInfo.customPosts)).
		Int("numNavigationLinks", len(websiteInfo.navigationLinks)).
		Int("numCategories", len(categories)).
		Int("numTags", len(tags)).
		Msgf("WebsiteInfo: %s", websiteInfo.title)
	return &websiteInfo, nil
}

func getAttachmentInfo(item *rss.Item, taxonomies []TaxonomyInfo) (*AttachmentInfo, error) {
	fields, err := getCommonFields(item, taxonomies)
	if err != nil {
		return nil, fmt.Errorf("error getting common fields: %w", err)
	}
	attachment := AttachmentInfo{*fields}
	log.Trace().
		Any("attachment", attachment).
		Msg("Attachment")
	return &attachment, nil
}

func getCommonFields(item *rss.Item, taxonomies []TaxonomyInfo) (*CommonFields, error) {
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
	pageTaxonomies := make([]TaxonomyInfo, 0, len(item.Categories))
	var postFormat *string

	for _, category := range item.Categories {
		if isCategory(category) {
			pageCategories = append(pageCategories, NormalizeCategoryName(category.Value))
		} else if isTag(category) {
			pageTags = append(pageTags, NormalizeCategoryName(category.Value))
		} else if isPostFormat(category) {
			tmp := NormalizeCategoryName(category.Value)
			postFormat = &tmp
		} else {
			taxo := isTaxonomy(category, taxonomies)
			if taxo != nil {
				pageTaxonomies = append(pageTaxonomies, *taxo)
			} else {
				log.Warn().
					Str("link", item.Link).
					Any("categories", item.Categories).
					Msgf("Unknown category: %s", category)
			}
		}
	}

	pageCustomMetaData := make([]CustomMetaDatum, 0, len(item.Extensions["wp"]["postmeta"]))
	// Extract custom metadata from <wp:postmeta>
	if len(item.Extensions["wp"]["postmeta"]) > 0 {
		for _, meta := range item.Extensions["wp"]["postmeta"] {
			var key, value string
			if len(meta.Children["meta_key"]) > 0 {
				key = meta.Children["meta_key"][0].Value
			}
			if len(meta.Children["meta_value"]) > 0 {
				value = meta.Children["meta_value"][0].Value
			}
			if key != "" {
				pageCustomMetaData = append(pageCustomMetaData, CustomMetaDatum{
					Key:   key,
					Value: value,
				})
			}
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
	if len(tmp1) > 1 {
		log.Warn().
			Str("link", item.Link).
			Any("attachmentURL", tmp1).
			Msg("Multiple attachment URLs")
	}

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

	var postType *string
	if len(item.Extensions["wp"]["post_type"]) > 0 {
		postType = &item.Extensions["wp"]["post_type"][0].Value
	} else {
		postType = nil
	}

	var postParent *string
	tmp := item.Extensions["wp"]["post_parent"][0].Value
	if tmp != "0" && tmp != "" {
		log.Debug().
			Str("link", item.Link).
			Str("post_parent", tmp).
			Msg("Item has a parent")
		postParent = &tmp
	} else {
		postParent = nil
	}

	comments := make([]CommentInfo, 0, len(item.Extensions["wp"]["comment"]))
	if len(item.Extensions["wp"]["comment"]) > 0 {
		for _, comment := range item.Extensions["wp"]["comment"] {
			// Don't append spams and unapproved comments
			if comment.Children["comment_approved"][0].Value == "1" {
				var commentPubDate *time.Time
				tmp, err := time.Parse("2006-01-02 15:04:05", comment.Children["comment_date"][0].Value)
				if err != nil {
					log.Warn().
						Str("date", item.Extensions["wp"]["post_date"][0].Value).
						Msg("Error parsing date")
				} else {
					pubDate = &tmp
				}

				comments = append(comments, CommentInfo{
					ID:          comment.Children["comment_id"][0].Value,
					ParentID:    comment.Children["comment_parent"][0].Value,
					AuthorName:  comment.Children["comment_author"][0].Value,
					AuthorEmail: comment.Children["comment_author_email"][0].Value,
					AuthorURL:   comment.Children["comment_author_url"][0].Value,
					PublishDate: commentPubDate,
					Content:     comment.Children["comment_content"][0].Value,
					PostLink:    item.Link,
					PostID:      item.Extensions["wp"]["post_id"][0].Value,
				})
			}
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
		PostType:         postType,
		PostParentID:     postParent,
		Excerpt:          item.Extensions["excerpt"]["encoded"][0].Value,

		Description:     item.Description,
		Content:         item.Content,
		Categories:      pageCategories,
		CustomMetaData:  pageCustomMetaData,
		Tags:            pageTags,
		Taxonomies:      pageTaxonomies,
		Footnotes:       getFootnotes(item),
		FeaturedImageID: getThumbnailID(item),

		attachmentURL: attachmentURL,

		Comments: comments,
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
	htmlCommentExtractor := regexp.MustCompile(`<!--(.*?)-->`)
	comments := htmlCommentExtractor.FindAllString(content, -1)
	log.Debug().
		Int("navigationLinks", len(comments)).
		Msg("getNavigationLinks")
	results := make([]NavigationLink, 0, len(comments))
	for _, comment := range comments {
		log.Trace().Msgf("comment: %s", comment)
		navigationLinkExtractor := regexp.MustCompile(`{.*}`)
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
	return category.Domain == "category" || category.Domain == "portfolio_category" || category.Domain == "product_cat"
}

func isPostFormat(category *rss.Category) bool {
	return category.Domain == "post_format"
}

func isTag(tag *rss.Category) bool {
	return tag.Domain == "post_tag" || tag.Domain == "portfolio_tags" || tag.Domain == "product_tag"
}

func isTaxonomy(taxonomy *rss.Category, taxonomies []TaxonomyInfo) *TaxonomyInfo {
	for _, tax := range taxonomies {
		if tax.Taxonomy == taxonomy.Domain {
			return &tax
		}
	}
	return nil
}

// NormalizeCategoryName removes space from the category name and converts it to lowercase
func NormalizeCategoryName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}

func getPageInfo(item *rss.Item, taxonomies []TaxonomyInfo) (*PageInfo, error) {
	fields, err := getCommonFields(item, taxonomies)
	if err != nil {
		return nil, fmt.Errorf("error getting common fields: %w", err)
	}
	page := PageInfo{*fields}
	log.Trace().
		Any("page", page).
		Msg("Page")
	return &page, nil
}

// testing only
func GetPostInfo(item *rss.Item, taxonomies []TaxonomyInfo) (*PostInfo, error) {
	return getPostInfo(item, taxonomies)
}

func getPostInfo(item *rss.Item, taxonomies []TaxonomyInfo) (*PostInfo, error) {
	fields, err := getCommonFields(item, taxonomies)
	if err != nil {
		return nil, fmt.Errorf("error getting common fields: %w", err)
	}
	post := PostInfo{*fields}
	log.Trace().
		Any("post", post).
		Msg("Post")
	return &post, nil
}

func getCustomPostInfo(item *rss.Item, taxonomies []TaxonomyInfo) (*CustomPostInfo, error) {
	fields, err := getCommonFields(item, taxonomies)
	if err != nil {
		return nil, fmt.Errorf("error getting common fields: %w", err)
	}
	post := CustomPostInfo{*fields}
	log.Trace().
		Any(*post.PostType, post).
		Msg("Custom Post")
	return &post, nil
}

func getCategories(inputs []ext.Extension) []CategoryInfo {
	categories := make([]CategoryInfo, 0, len(inputs))
	for _, input := range inputs {
		categoryName := ""
		if len(input.Children["cat_name"]) > 0 {
			categoryName = NormalizeCategoryName(input.Children["cat_name"][0].Value)
		}
		categoryNiceName := ""
		if len(input.Children["category_nicename"]) > 0 {
			categoryNiceName = input.Children["category_nicename"][0].Value
		}
		category := CategoryInfo{
			// ID is usually int but for safety let's assume string
			ID:       input.Children["term_id"][0].Value,
			Name:     categoryName,
			NiceName: categoryNiceName,
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

func buildTaxonomy(term ext.Extension) TaxonomyInfo {
	var id int
	var taxonomy, slug, parent, name string

	if len(term.Children["term_slug"]) > 0 {
		slug = term.Children["term_slug"][0].Value
	}
	if len(term.Children["term_name"]) > 0 {
		name = term.Children["term_name"][0].Value
	}
	if len(term.Children["term_parent"]) > 0 {
		parent = term.Children["term_parent"][0].Value
	}
	if len(term.Children["term_taxonomy"]) > 0 {
		taxonomy = term.Children["term_taxonomy"][0].Value
	}
	if len(term.Children["term_id"]) > 0 {
		idStr := term.Children["term_id"][0].Value
		var err error
		id, err = strconv.Atoi(idStr)
		if err != nil {
			log.Warn().
				Str("term_id", idStr).
				Msg("Error converting term_id to int")
			id = 0
		}
	}
	return TaxonomyInfo{
		ID:       id,
		Taxonomy: taxonomy,
		Parent:   parent,
		Name:     name,
		Slug:     slug,
	}
}

func getTaxonomies(inputs []ext.Extension) []TaxonomyInfo {
	taxonomies := make([]TaxonomyInfo, 0, len(inputs))
	for _, term := range inputs {
		taxonomies = append(taxonomies, buildTaxonomy(term))
	}
	return taxonomies
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
