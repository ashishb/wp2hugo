package hugopage

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	_sampleHTMLInput1 = `
<!-- wp:paragraph -->
<p><a href="https://example.com">Hello world</a>. "<a href="https://example.com">Example 1</a>".</p>
<!-- /wp:paragraph -->
`
	_sampleMarkdownOutput1 = `[Hello world](https://example.com). "[Example 1](https://example.com)".`
)

const (
	_sampleHTMLInput2      = `Unlike <a href="https://some.com/link1">his</a> <a href="https://some.com/link2">other</a>, this book`
	_sampleMarkdownOutput2 = `Unlike [his](https://some.com/link1) [other](https://some.com/link2), this book`
)

const (
	_sampleHTMLInput3      = `<ol><li>First item</li><li>Second item</li><li>Third item</li></ol>`
	_sampleMarkdownOutput3 = "1. First item\n1. Second item\n1. Third item"
)

const (
	_sampleHTMLInput4      = `This is<br><br>some<br><br><br>tExt`
	_sampleMarkdownOutput4 = "This is\n\nsome\n\ntExt"
)

const (
	_sampleHTMLInput5      = "<!-- wp:paragraph --><p>First line<br /><abcedef>Second line<br>Third line</p><!-- /wp:paragraph -->"
	_sampleMarkdownOutput5 = "First line  \nSecond line  \nThird line"
)

func TestMarkdownExtractorWithLink1(t *testing.T) {
	testMarkdownExtractor(t, _sampleHTMLInput2, _sampleMarkdownOutput2)
}

func TestMarkdownExtractorWithLink2(t *testing.T) {
	// Ref:
	// 1. https://github.com/ashishb/wp2hugo/issues/11
	// 2. https://github.com/JohannesKaufmann/html-to-markdown/issues/95
	t.Skipf("This is failing due to a bug in the underlying library. Skipping for now.")
	testMarkdownExtractor(t, _sampleHTMLInput1, _sampleMarkdownOutput1)
}

func TestListExtractor(t *testing.T) {
	testMarkdownExtractor(t, _sampleHTMLInput3, _sampleMarkdownOutput3)
}

func TestConsecutiveNewlines(t *testing.T) {
	testMarkdownExtractor(t, _sampleHTMLInput4, _sampleMarkdownOutput4)
}

func TestManualLineBreaks(t *testing.T) {
	testMarkdownExtractor(t, _sampleHTMLInput5, _sampleMarkdownOutput5)
}

func testMarkdownExtractor(t *testing.T, htmlInput string, markdownOutput string) {
	url1, err := url.Parse("https://example.com")
	assert.Nil(t, err)
	page, err := NewPage(nil, *url1, "author", "Title", nil, false, nil, nil, nil, nil, htmlInput, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	md, err := page.getMarkdown(nil, htmlInput, nil)
	assert.NoError(t, err)
	assert.Equal(t, markdownOutput, *md)
}

func TestPreTagExtractor2(t *testing.T) {
	const example1 = `<pre class="lang:js decode:true">document.querySelector("video").playbackRate = 2.0;   // For 2X speed-up</pre>`
	const example2 = `<pre class="theme:solarized-dark lang:sh decode:true">echo "whatever"</pre>`
	const example3 = `<pre class="lang:sh decode:true"># Sample invocation:\n</pre>`
	assert.True(t, _preTagExtractor2.MatchString(example1))
	assert.True(t, _preTagExtractor2.MatchString(example2))
	assert.True(t, _preTagExtractor2.MatchString(example3))

	result3 := _preTagExtractor2.FindAllStringSubmatch(example3, -1)
	assert.Len(t, result3, 1)
	assert.Len(t, result3[0], 3)
	assert.Equal(t, "sh", result3[0][1])
}
