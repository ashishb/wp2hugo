package hugopage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProtectAndRestoreCode(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name              string
		input             string
		expectedProtected string
	}{
		{
			name:              "fenced code block",
			input:             "Text\n\n```go\na()\n\n\nb()\n```\n\nMore",
			expectedProtected: "Text\n\n\x00wp2hugo-code-0\x00\n\nMore",
		},
		{
			name:              "tilde fence and longer closing fence",
			input:             "~~~\nx\n~~~~\nafter",
			expectedProtected: "\x00wp2hugo-code-0\x00\nafter",
		},
		{
			name:              "a fence with an info string does not close the block",
			input:             "````\n```go\nx\n```\n````\nafter",
			expectedProtected: "\x00wp2hugo-code-0\x00\nafter",
		},
		{
			name:              "indented code block in a list item keeps its indentation",
			input:             "1. Step\n\n   ```\n   x\n   ```\n\n1. Next",
			expectedProtected: "1. Step\n\n   \x00wp2hugo-code-0\x00\n\n1. Next",
		},
		{
			name:              "unclosed code block runs until the end",
			input:             "a\n```\nb\n\nc",
			expectedProtected: "a\n\x00wp2hugo-code-0\x00",
		},
		{
			name:              "inline code",
			input:             "Run `curl https://example.com` and ``a ` b`` now",
			expectedProtected: "Run \x00wp2hugo-code-0\x00 and \x00wp2hugo-code-1\x00 now",
		},
		{
			// The HTML to Markdown converter escapes backticks that are part of the text
			name:              "escaped backticks are not code",
			input:             "Use \\` here and \\`not code\\` there",
			expectedProtected: "Use \\` here and \\`not code\\` there",
		},
		{
			name:              "unmatched backtick is not code",
			input:             "a `b c",
			expectedProtected: "a `b c",
		},
		{
			name:              "inline code does not span paragraphs",
			input:             "a `b\n\nc` d",
			expectedProtected: "a `b\n\nc` d",
		},
		{
			name:              "inline code after a fenced code block",
			input:             "```\nx\n```\nSee `y`",
			expectedProtected: "\x00wp2hugo-code-0\x00\nSee \x00wp2hugo-code-1\x00",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			protected, code := protectCode(tc.input)
			require.Equal(t, tc.expectedProtected, protected)
			require.Equal(t, tc.input, restoreCode(protected, code))
		})
	}
}

func TestProtectBlankLinesInPre(t *testing.T) {
	t.Parallel()
	const marker = _blankLineInCodeMarker
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "blank lines between code lines",
			input:    "<pre><code>a\n\n\nb\n  \nc</code></pre>",
			expected: "<pre><code>a\n" + marker + "\n" + marker + "\nb\n" + marker + "\nc</code></pre>",
		},
		{
			name:     "leading and trailing blank lines are kept",
			input:    "<pre class=\"x\">\n\na\n\nb\n\n</pre>",
			expected: "<pre class=\"x\">\n\na\n" + marker + "\nb\n\n</pre>",
		},
		{
			name:     "text outside of pre is not touched",
			input:    "<p>a</p>\n\n\n<pre>x</pre>\n\n\n<p>b</p>",
			expected: "<p>a</p>\n\n\n<pre>x</pre>\n\n\n<p>b</p>",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, protectBlankLinesInPre(tc.input))
		})
	}
}

func TestCodeIsNotModifiedByMarkdownPostProcessing(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name         string
		htmlInput    string
		expectedBody string
	}{
		{
			name:         "absolute links in code are kept",
			htmlInput:    `<p>See <a href="https://example.com/about">about</a></p><pre><code>curl https://example.com/api/v1</code></pre><p>Run <code>curl https://example.com/api</code></p>`,
			expectedBody: "See [about](/about)\n\n```\ncurl https://example.com/api/v1\n```\n\nRun `curl https://example.com/api`",
		},
		{
			name:         "consecutive blank lines in code are kept",
			htmlInput:    "<p>a</p>\n\n\n\n<pre><code class=\"language-python\">import os\n\n\ndef f():\n    pass</code></pre>",
			expectedBody: "a\n\n```python\nimport os\n\n\ndef f():\n    pass\n```",
		},
		{
			name:         "YouTube URLs in code are not converted",
			htmlInput:    "<pre><code>https://www.youtube.com/watch?v=gJ7AAJXHeeg</code></pre><p>https://youtu.be/gJ7AAJXHeeg</p>",
			expectedBody: "```\nhttps://www.youtube.com/watch?v=gJ7AAJXHeeg\n```\n\n{{< youtube gJ7AAJXHeeg >}}",
		},
		{
			name:         "spaces before brackets in code are kept",
			htmlInput:    "<pre><code>print( [1, 2])\nx = \" [a]\"</code></pre><p>(<a href=\"https://x.org\">x</a>)</p>",
			expectedBody: "```\nprint( [1, 2])\nx = \" [a]\"\n```\n\n([x](https://x.org))",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testMarkdownExtractor(t, tc.htmlInput, tc.expectedBody)
		})
	}
}
