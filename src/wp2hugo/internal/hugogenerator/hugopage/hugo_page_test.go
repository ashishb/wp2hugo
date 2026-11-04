package hugopage

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// Ref: https://github.com/ashishb/wp2hugo/issues/11
	_sampleHTMLInput1 = `
<!-- wp:paragraph -->
<p><a href="https://example.com">Hello world</a>. "<a href="https://example.com">Example 1</a>". (<a href="https://example.com">Example 2</a>)</p>
<!-- /wp:paragraph -->
`
	_sampleMarkdownOutput1 = `[Hello world](https://example.com). "[Example 1](https://example.com)". ([Example 2](https://example.com))`
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

const (
	// Ref: https://github.com/ashishb/wp2hugo/issues/11 (single-quote case)
	_sampleHTMLInput6      = `<p>'<a href="https://example.com">Example</a>'</p>`
	_sampleMarkdownOutput6 = `'[Example](https://example.com)'`
)

func TestMarkdownExtractorWithLink1(t *testing.T) {
	t.Parallel()
	testMarkdownExtractor(t, _sampleHTMLInput2, _sampleMarkdownOutput2)
}

func TestMarkdownExtractorWithLink2(t *testing.T) {
	t.Parallel()
	// Ref:
	// 1. https://github.com/ashishb/wp2hugo/issues/11
	// 2. https://github.com/JohannesKaufmann/html-to-markdown/issues/95
	testMarkdownExtractor(t, _sampleHTMLInput1, _sampleMarkdownOutput1)
}

func TestMarkdownExtractorWithLink3(t *testing.T) {
	t.Parallel()
	// Ref: https://github.com/ashishb/wp2hugo/issues/11 (single-quote case)
	testMarkdownExtractor(t, _sampleHTMLInput6, _sampleMarkdownOutput6)
}

func TestListExtractor(t *testing.T) {
	t.Parallel()
	testMarkdownExtractor(t, _sampleHTMLInput3, _sampleMarkdownOutput3)
}

func TestListWithStartAttributeKeepsNumbers(t *testing.T) {
	t.Parallel()
	const htmlInput = `<ol><li>One</li><li>Two</li></ol><p>Picture here</p>` +
		`<ol start="4"><li>Four</li><li>Five<ol><li>Sub a</li><li>Sub b</li></ol></li><li>Six</li></ol>`
	const markdownOutput = "1. One\n1. Two\n\nPicture here\n\n4. Four\n5. Five\n   1. Sub a\n   1. Sub b\n6. Six"
	testMarkdownExtractor(t, htmlInput, markdownOutput)
}

func TestListNumbersInsideCodeBlockAreUntouched(t *testing.T) {
	t.Parallel()
	const htmlInput = "<pre><code>version: 3\n2. second line\n10. tenth line</code></pre><ol><li>Item</li><li>Item</li></ol>"
	const markdownOutput = "```\nversion: 3\n2. second line\n10. tenth line\n```\n\n1. Item\n1. Item"
	testMarkdownExtractor(t, htmlInput, markdownOutput)
}

func TestListItemWithLineBreak(t *testing.T) {
	t.Parallel()
	const htmlInput = `<ol><li>First<br>second line</li><li>Third</li></ol>`
	const markdownOutput = "1. First  \nsecond line\n1. Third"
	testMarkdownExtractor(t, htmlInput, markdownOutput)
}

func TestReplaceOrderedListNumbers(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"list starting at 1", "1. a\n2. b\n3. c", "1. a\n1. b\n1. c"},
		{"list starting at 3", "3. a\n4. b", "3. a\n4. b"},
		{"loose list", "1. a\n\n2. b", "1. a\n\n1. b"},
		{"continuation paragraph", "1. a\n\n   more text\n\n2. b", "1. a\n\n   more text\n\n1. b"},
		{"lazy continuation line", "1. a  \nsecond line\n2. b\n3. c", "1. a  \nsecond line\n1. b\n1. c"},
		{"lazy continuation in list starting at 4", "4. a\nlazy\n5. b", "4. a\nlazy\n5. b"},
		{"paragraph ends the list", "5. a\n6. b\n\ntext\n\n1. c\n2. d", "5. a\n6. b\n\ntext\n\n1. c\n1. d"},
		{"tilde fence", "~~~\n2. x\n~~~\n2. y\n3. z", "~~~\n2. x\n~~~\n2. y\n3. z"},
		{"backtick fence", "```go\n7. x\n```\n1. y\n2. z", "```go\n7. x\n```\n1. y\n1. z"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, replaceOrderedListNumbers(tc.input))
		})
	}
}

func TestConsecutiveNewlines(t *testing.T) {
	t.Parallel()
	testMarkdownExtractor(t, _sampleHTMLInput4, _sampleMarkdownOutput4)
}

func TestManualLineBreaks(t *testing.T) {
	t.Parallel()
	testMarkdownExtractor(t, _sampleHTMLInput5, _sampleMarkdownOutput5)
}

func testMarkdownExtractor(t *testing.T, htmlInput string, markdownOutput string) {
	t.Helper()
	url1, err := url.Parse("https://example.com")
	require.NoError(t, err)
	page, err := NewPage(nil, *url1, "author", "Title", nil, false, nil, nil, nil, nil, htmlInput, nil, nil, nil, nil, nil, "0", nil)
	require.NoError(t, err)
	md, err := page.getMarkdown(nil, htmlInput, nil)
	require.NoError(t, err)
	require.Equal(t, markdownOutput, *md)
}

func TestPreTagExtractor2(t *testing.T) {
	t.Parallel()
	const example1 = `<pre class="lang:js decode:true">document.querySelector("video").playbackRate = 2.0;   // For 2X speed-up</pre>`
	const example2 = `<pre class="theme:solarized-dark lang:sh decode:true">echo "whatever"</pre>`
	const example3 = `<pre class="lang:sh decode:true"># Sample invocation:\n</pre>`
	require.True(t, _preTagExtractor2.MatchString(example1))
	require.True(t, _preTagExtractor2.MatchString(example2))
	require.True(t, _preTagExtractor2.MatchString(example3))

	result3 := _preTagExtractor2.FindAllStringSubmatch(example3, -1)
	require.Len(t, result3, 1)
	require.Len(t, result3[0], 3)
	require.Equal(t, "sh", result3[0][1])
}
