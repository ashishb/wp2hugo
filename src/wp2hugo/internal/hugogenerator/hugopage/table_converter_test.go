package hugopage

import (
	"testing"
)

func TestTableConversion(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name         string
		htmlInput    string
		expectedBody string
	}{
		{
			name: "Gutenberg table block with header and caption",
			htmlInput: `<!-- wp:table --><figure class="wp-block-table"><table><thead><tr><th>Name</th><th>Price</th></tr></thead>` +
				`<tbody><tr><td>Apple</td><td>$1</td></tr><tr><td>Pear</td><td>$2</td></tr></tbody></table>` +
				`<figcaption class="wp-element-caption">Fruit prices</figcaption></figure><!-- /wp:table -->`,
			expectedBody: "| Name | Price |\n| --- | --- |\n| Apple | $1 |\n| Pear | $2 |\n\nFruit prices",
		},
		{
			name:         "table without a header row gets an empty header",
			htmlInput:    "<table>\n<tr>\n<td>a</td>\n<td>b</td>\n</tr>\n<tr><td>c</td><td>d</td></tr>\n</table>\n\nAfter",
			expectedBody: "|     |     |\n| --- | --- |\n| a | b |\n| c | d |\n\nAfter",
		},
		{
			name:         "column alignment",
			htmlInput:    `<table><thead><tr><th>L</th><th align="right">R</th><th align="center">C</th></tr></thead><tbody><tr><td>1</td><td>2</td><td>3</td></tr></tbody></table>`,
			expectedBody: "| L | R | C |\n| --- | --: | :-: |\n| 1 | 2 | 3 |",
		},
		{
			name:         "line breaks in a cell stay in the row",
			htmlInput:    `<table><tr><th>H</th></tr><tr><td>line1<br>line2<br/>line3</td></tr></table>`,
			expectedBody: "| H |\n| --- |\n| line1<br>line2<br>line3 |",
		},
		{
			name:         "paragraphs in a cell",
			htmlInput:    `<table><tr><th>H</th></tr><tr><td><p>a</p><p>b</p></td></tr></table>`,
			expectedBody: "| H |\n| --- |\n| a<br>b |",
		},
		{
			name:         "links, formatting and pipes in a cell",
			htmlInput:    `<table><tr><th>H</th><th>I</th></tr><tr><td><a href="https://example.com/x">link</a> <strong>bold</strong></td><td>a | b</td></tr></table>`,
			expectedBody: "| H | I |\n| --- | --- |\n| [link](/x) **bold** | a \\| b |",
		},
		{
			name:         "pipes in inline code in a cell are escaped",
			htmlInput:    "<table><tr><th>Command</th></tr><tr><td>Run <code>ls | wc -l</code> now</td></tr></table>",
			expectedBody: "| Command |\n| --- |\n| Run `ls \\| wc -l` now |",
		},
		{
			name:         "YouTube URLs in a cell are not embedded",
			htmlInput:    "<table><tr><th>Video</th></tr><tr><td>https://www.youtube.com/watch?v=gJ7AAJXHeeg</td></tr></table><p>https://youtu.be/gJ7AAJXHeeg</p>",
			expectedBody: "| Video |\n| --- |\n| https://www.youtube.com/watch?v=gJ7AAJXHeeg |\n\n{{< youtube gJ7AAJXHeeg >}}",
		},
		{
			name:         "table in a list item and in a blockquote",
			htmlInput:    "<ul><li>Item<table><tr><th>A</th></tr><tr><td>1</td></tr></table></li><li>Next</li></ul><blockquote><table><tr><th>B</th></tr><tr><td>https://youtu.be/gJ7AAJXHeeg</td></tr></table></blockquote>",
			expectedBody: "- Item\n\n  | A |\n  | --- |\n  | 1 |\n\n- Next\n\n> | B |\n> | --- |\n> | https://youtu.be/gJ7AAJXHeeg |",
		},
		{
			name:         "line breaks and inline code outside of a table are unchanged",
			htmlInput:    "<p>one<br>two <code>a | b</code></p>",
			expectedBody: "one  \ntwo `a | b`",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testMarkdownExtractor(t, tc.htmlInput, tc.expectedBody)
		})
	}
}

func TestLongestBacktickRun(t *testing.T) {
	t.Parallel()
	testCases := map[string]int{
		"":         0,
		"abc":      0,
		"a`b":      1,
		"a``b`c":   2,
		"```":      3,
		"x`y```z`": 3,
	}
	for input, expected := range testCases {
		if got := longestBacktickRun(input); got != expected {
			t.Errorf("longestBacktickRun(%q) = %d, want %d", input, got, expected)
		}
	}
}
