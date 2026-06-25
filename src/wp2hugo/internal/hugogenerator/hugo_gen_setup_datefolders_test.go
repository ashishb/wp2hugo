package hugogenerator

import (
	"path"
	"testing"
	"time"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/stretchr/testify/require"
)

func TestGetDateBasedContentDir(t *testing.T) {
	t.Parallel()
	publishDate := time.Date(2025, time.November, 2, 1, 2, 3, 0, time.UTC)

	baseDir := "/tmp/site/content/posts"
	require.Equal(t, baseDir, getDateBasedContentDir(baseDir, &publishDate, ContentDateFolderStructureFlat))
	require.Equal(t, path.Join(baseDir, "2025"), getDateBasedContentDir(baseDir, &publishDate, ContentDateFolderStructureYear))
	require.Equal(t, path.Join(baseDir, "2025", "11"), getDateBasedContentDir(baseDir, &publishDate, ContentDateFolderStructureYearMonth))
	require.Equal(t, baseDir, getDateBasedContentDir(baseDir, nil, ContentDateFolderStructureYearMonth))
}

func TestGetPagePath_WithYearMonthFolders_NoParent(t *testing.T) {
	t.Parallel()
	outputDir := t.TempDir()
	postType := "page"
	publishDate := time.Date(2026, time.March, 4, 10, 0, 0, 0, time.UTC)

	pagePath, err := getPagePath(outputDir, wpparser.CommonFields{
		Title:       "My page",
		Link:        "https://example.com/my-page/",
		PostType:    &postType,
		PublishDate: &publishDate,
	}, nil, ContentDateFolderStructureYearMonth)
	require.NoError(t, err)
	require.Equal(t, path.Join(outputDir, "content", "pages", "2026", "03", "my-page", "_index.md"), pagePath)
}

func TestGetPagePath_WithYearMonthFolders_ChildUsesParentFolder(t *testing.T) {
	t.Parallel()
	outputDir := t.TempDir()
	postType := "page"
	parentID := "10"
	childParentID := "10"
	parentDate := time.Date(2025, time.January, 8, 9, 0, 0, 0, time.UTC)
	childDate := time.Date(2026, time.April, 10, 9, 0, 0, 0, time.UTC)

	parent := wpparser.CommonFields{
		PostID:       parentID,
		Title:        "Parent Page",
		Link:         "https://example.com/parent-page/",
		PostType:     &postType,
		PublishDate:  &parentDate,
		PostParentID: nil,
	}
	child := wpparser.CommonFields{
		Title:        "Child Page",
		Link:         "https://example.com/child-page/",
		PostType:     &postType,
		PublishDate:  &childDate,
		PostParentID: &childParentID,
	}

	pagePath, err := getPagePath(outputDir, child, []wpparser.CommonFields{parent}, ContentDateFolderStructureYearMonth)
	require.NoError(t, err)
	require.Equal(t, path.Join(outputDir, "content", "pages", "2025", "01", "parent-page", "child-page.md"), pagePath)
}
