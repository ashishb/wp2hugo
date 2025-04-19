package hugogenerator

import (
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/stretchr/testify/require"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestFootnote(t *testing.T) {
	file, err := os.Open("./testdata/testcase.WordPress.2024-07-01.xml")
	require.NoError(t, err)

	parser := wpparser.NewParser()
	websiteInfo, err := parser.Parse(file, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(websiteInfo.Posts()))

	post := websiteInfo.Posts()[0]
	require.Equal(t, 1, len(post.Footnotes))
	require.NotNil(t, post.CommonFields)

	generator := NewGenerator("/tmp", "", nil, false, false, true, *websiteInfo)
	url1, err := url.Parse(post.GUID.Value)
	require.NoError(t, err)
	hugoPage, err := generator.newHugoPage(url1, post.CommonFields)
	require.NoError(t, err)

	const expectedMarkdown = "Some text[^1] with a footnote\n\n[^1]: Here we are: the footnote."
	require.True(t, strings.Contains(hugoPage.Markdown(), expectedMarkdown))
}

func TestPost(t *testing.T) {
	file, err := os.Open("./testdata/testcase.WordPress_2.xml")
	require.NoError(t, err)

	parser := wpparser.NewParser()
	websiteInfo, err := parser.Parse(file, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(websiteInfo.Posts()))

	post := websiteInfo.Posts()[0]
	require.NotNil(t, post.CommonFields)
	require.Equal(t, "Kurz angemerkt zum Tag der Schachtels��tze", post.Title)
	require.Equal(t, 1, len(post.Categories))
	require.Equal(t, "netzfundst��cke", post.Categories[0])
	require.Equal(t, 1276, len(post.Content))
}
