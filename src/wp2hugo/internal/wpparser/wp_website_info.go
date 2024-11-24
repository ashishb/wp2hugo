package wpparser

import (
	"time"
)

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
	CustomPosts     []CustomPostInfo

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

func (w *WebsiteInfo) GetAttachmentsForPost(postID string) []AttachmentInfo {
	return w.postIDToAttachmentCache[postID]
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
