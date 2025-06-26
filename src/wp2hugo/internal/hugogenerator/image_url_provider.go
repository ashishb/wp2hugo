package hugogenerator

import (
	"fmt"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugogenerator/hugopage"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/rs/zerolog/log"
)

type WordPressImageURLProvider struct {
	info wpparser.WebsiteInfo
}

func (w WordPressImageURLProvider) GetImageInfo(imageID int) (*hugopage.ImageInfo, error) {
	log.Debug().
		Int("imageID", imageID).
		Msg("GetImageURL")
	for _, attachment := range w.info.Attachments() {
		if attachment.PostID == imageID {
			attachmentURL := attachment.GetAttachmentURL()
			if attachmentURL != nil {
				log.Info().
					Int("imageID", imageID).
					Str("Link", *attachmentURL).
					Msg("Image URL found")
				return &hugopage.ImageInfo{
					ImageURL: *attachmentURL,
					Title:    attachment.Title,
				}, nil
			}
		}
	}
	log.Error().Int("imageID", imageID).Msg("Image URL not found")
	return nil, fmt.Errorf("image URL not found for imageID: %s", imageID)
}

func newImageURLProvider(info wpparser.WebsiteInfo) WordPressImageURLProvider {
	return WordPressImageURLProvider{info: info}
}
