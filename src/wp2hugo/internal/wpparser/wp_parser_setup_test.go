package wpparser

import (
	"testing"

	ext "github.com/mmcdole/gofeed/extensions"
	"github.com/mmcdole/gofeed/rss"
	"github.com/stretchr/testify/require"
)

func TestGetCommonFields_UnknownPublishStatusFallsBackToDraft(t *testing.T) {
	t.Parallel()

	fields, err := getCommonFields(newRSSItemWithStatus("mystery"), nil)
	require.NoError(t, err)
	require.NotNil(t, fields)
	require.Equal(t, PublishStatusDraft, fields.PublishStatus)
}

func TestGetCommonFields_TrashReturnsIgnoredError(t *testing.T) {
	t.Parallel()

	fields, err := getCommonFields(newRSSItemWithStatus(string(PublishStatusTrash)), nil)
	require.Nil(t, fields)
	require.Error(t, err)
	require.ErrorIs(t, err, errTrashItem)
}

func newRSSItemWithStatus(status string) *rss.Item {
	return &rss.Item{
		Title:       "test-title",
		Link:        "https://example.com/test-title",
		Description: "desc",
		Content:     "content",
		Extensions: map[string]map[string][]ext.Extension{
			"wp": {
				"status":      {{Value: status}},
				"post_id":     {{Value: "1"}},
				"post_parent": {{Value: "0"}},
			},
			"excerpt": {
				"encoded": {{Value: ""}},
			},
		},
	}
}
