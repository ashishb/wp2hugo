package wpparser

import (
	"net/url"
	"time"
)

type WebsiteInfo struct {
	title       string
	link        *url.URL
	Description string

	pubDate  *time.Time
	language string

	categories []CategoryInfo
	tags       []TagInfo

	// Collecting attachments is mostly useless, but we are doing it for completeness
	// Only the ones that are actually used in posts/pages are useful
	attachments     []AttachmentInfo
	pages           []PageInfo
	posts           []PostInfo
	navigationLinks []NavigationLink
	customPosts     []CustomPostInfo

	postIDToAttachmentCache map[string][]AttachmentInfo
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

func (w *WebsiteInfo) Title() string {
	return w.title
}

func (w *WebsiteInfo) Link() *url.URL {
	return w.link
}

func (w *WebsiteInfo) Language() string {
	return w.language
}

func (w *WebsiteInfo) Attachments() []AttachmentInfo {
	return w.attachments
}

func (w *WebsiteInfo) NavigationLinks() []NavigationLink {
	return w.navigationLinks
}

func (w *WebsiteInfo) GetAttachmentsForPost(postID string) []AttachmentInfo {
	return w.postIDToAttachmentCache[postID]
}

func (w *WebsiteInfo) Pages() []PageInfo {
	return w.pages
}

func (w *WebsiteInfo) Posts() []PostInfo {
	return w.posts
}

func (w *WebsiteInfo) CustomPosts() []CustomPostInfo {
	return w.customPosts
}

func getPostIDToAttachmentsMap(attachments []AttachmentInfo) map[string][]AttachmentInfo {
	result := make(map[string][]AttachmentInfo)
	for _, attachment := range attachments {
		if attachment.PostParentID == nil {
			continue
		}
		parentID := *attachment.PostParentID
		if result[parentID] == nil {
			result[parentID] = make([]AttachmentInfo, 0, 1)
		}
		result[parentID] = append(result[parentID], attachment)
	}
	return result
}
