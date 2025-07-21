package hugogenerator

import (
	"net/url"
	"os"
	"testing"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/stretchr/testify/require"
)

func TestFootnote(t *testing.T) {
	t.Parallel()
	file, err := os.Open("./testdata/testcase.WordPress.2024-07-01.xml")
	require.NoError(t, err)

	parser := wpparser.NewParser()
	websiteInfo, err := parser.Parse(file, nil, nil)
	require.NoError(t, err)
	require.Len(t, websiteInfo.Posts(), 1)

	post := websiteInfo.Posts()[0]
	require.Len(t, post.Footnotes, 1)
	require.NotNil(t, post.CommonFields)

	generator := NewGenerator("/tmp", "", nil, false, false, false, true, *websiteInfo)
	url1, err := url.Parse(post.GUID.Value)
	require.NoError(t, err)
	hugoPage, err := generator.newHugoPage(url1, post.CommonFields)
	require.NoError(t, err)

	const expectedMarkdown = "Some text[^1] with a footnote\n\n[^1]: Here we are: the footnote."
	require.Contains(t, hugoPage.Markdown(), expectedMarkdown)
}

func TestPost(t *testing.T) {
	t.Parallel()
	file, err := os.Open("./testdata/testcase.WordPress_2.xml")
	require.NoError(t, err)

	parser := wpparser.NewParser()
	websiteInfo, err := parser.Parse(file, nil, nil)
	require.NoError(t, err)
	require.Len(t, websiteInfo.Posts(), 1)

	post := websiteInfo.Posts()[0]
	require.NotNil(t, post.CommonFields)
	require.Equal(t, "Kurz angemerkt zum Tag der Schachtels��tze", post.Title)
	require.Len(t, post.Categories, 1)
	require.Equal(t, "netzfundst��cke", post.Categories[0])
	require.Len(t, post.Content, 1276)
}
