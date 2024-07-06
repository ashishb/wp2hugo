package hugogenerator

import (
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/stretchr/testify/assert"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestFootnote(t *testing.T) {
	file, err := os.Open("./testdata/testcase.WordPress.2024-07-01.xml")
	assert.NoError(t, err)

	parser := wpparser.NewParser()
	websiteInfo, err := parser.Parse(file)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(websiteInfo.Posts))

	post := websiteInfo.Posts[0]
	assert.Equal(t, 1, len(post.Footnotes))
	assert.NotNil(t, post.CommonFields)

	generator := NewGenerator("/tmp", false, "", nil, *websiteInfo)
	url1, err := url.Parse(post.GUID.Value)
	assert.NoError(t, err)
	hugoPage, err := generator.NewHugoPage(url1, post.CommonFields)
	assert.NoError(t, err)

	const expectedMarkdown = "Some text[^1] with a footnote\n\n[^1]: Here we are: the footnote."
	assert.True(t, strings.Contains(hugoPage.Markdown(), expectedMarkdown))
}
