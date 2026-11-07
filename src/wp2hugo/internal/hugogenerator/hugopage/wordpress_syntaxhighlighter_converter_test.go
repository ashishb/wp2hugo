package hugopage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplaceSyntaxHighlighterShortcodes(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "language attribute",
			input:    "[code language=\"python\"]\ndef f(x):\n    return x\n[/code]",
			expected: "<pre><code class=\"language-python\">def f(x):\n    return x</code></pre>",
		},
		{
			name:     "lang attribute with single quotes",
			input:    "[sourcecode lang='bash']echo hi[/sourcecode]",
			expected: "<pre><code class=\"language-bash\">echo hi</code></pre>",
		},
		{
			name:     "unquoted attribute and other attributes",
			input:    "[code highlight=\"2\" lang=php light=\"true\"]<?php echo 1; ?>[/code]",
			expected: "<pre><code class=\"language-php\">&lt;?php echo 1; ?&gt;</code></pre>",
		},
		{
			name:     "no language",
			input:    "[code]x = 1[/code]",
			expected: "<pre><code>x = 1</code></pre>",
		},
		{
			name:     "shortcode named after the language",
			input:    "[python]\nprint('a')\n[/python]",
			expected: "<pre><code class=\"language-python\">print(&#39;a&#39;)</code></pre>",
		},
		{
			name:     "brush alias mapped to a Chroma language",
			input:    "[jscript]var a = 1;[/jscript]",
			expected: "<pre><code class=\"language-javascript\">var a = 1;</code></pre>",
		},
		{
			name:     "plain text brush has no language",
			input:    "[plain]just text[/plain]",
			expected: "<pre><code>just text</code></pre>",
		},
		{
			name:     "uppercase shortcode",
			input:    "[CODE LANGUAGE=\"Python\"]pass[/CODE]",
			expected: "<pre><code class=\"language-python\">pass</code></pre>",
		},
		{
			name:     "escaped code from the visual editor is not double escaped",
			input:    "[code lang=\"html\"]&lt;div&gt;&amp;&lt;/div&gt;[/code]",
			expected: "<pre><code class=\"language-html\">&lt;div&gt;&amp;&lt;/div&gt;</code></pre>",
		},
		{
			name:     "raw HTML from the text editor is escaped",
			input:    "[code lang=\"html\"]<div class=\"a\">x</div>[/code]",
			expected: "<pre><code class=\"language-html\">&lt;div class=&#34;a&#34;&gt;x&lt;/div&gt;</code></pre>",
		},
		{
			name:     "CRLF line endings",
			input:    "[code lang=\"go\"]\r\nfunc main() {\r\n\tfmt.Println()\r\n}\r\n[/code]",
			expected: "<pre><code class=\"language-go\">func main() {\n\tfmt.Println()\n}</code></pre>",
		},
		{
			name:     "multiple shortcodes",
			input:    "Before [code lang=\"go\"]a()[/code] middle [bash]b[/bash] after",
			expected: "Before <pre><code class=\"language-go\">a()</code></pre> middle <pre><code class=\"language-bash\">b</code></pre> after",
		},
		{
			name:     "tag without a matching closing tag is kept",
			input:    "Use the [c] and [cpp] tags or [code]x[/sourcecode]",
			expected: "Use the [c] and [cpp] tags or [code]x[/sourcecode]",
		},
		{
			name:     "escaped shortcode with content is kept as text",
			input:    "Example: [[code language=\"php\"]echo 1;[/code]] renders as code.",
			expected: "Example: [code language=\"php\"]echo 1;[/code] renders as code.",
		},
		{
			name:     "escaped opening tag is kept as text",
			input:    "Wrap your code in [[python]] and [/python]: [python]x = 1[/python]",
			expected: "Wrap your code in [python] and [/python]: <pre><code class=\"language-python\">x = 1</code></pre>",
		},
		{
			name:     "single bracket before a shortcode is not an escape",
			input:    "[[code]x[/code]",
			expected: "[<pre><code>x</code></pre>",
		},
		{
			name:     "other shortcodes are not touched",
			input:    "[caption id=\"1\"]x[/caption] [codepen id=\"2\"] [python_docs]",
			expected: "[caption id=\"1\"]x[/caption] [codepen id=\"2\"] [python_docs]",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, replaceSyntaxHighlighterCode(tc.input))
		})
	}
}

func TestReplaceSyntaxHighlighterGutenbergBlock(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "with language",
			input: "<!-- wp:syntaxhighlighter/code {\"language\":\"python\"} -->\n" +
				"<pre class=\"wp-block-syntaxhighlighter-code\">if x &lt; 1:\n    return x</pre>\n" +
				"<!-- /wp:syntaxhighlighter/code -->",
			expected: "<pre><code class=\"language-python\">if x &lt; 1:\n    return x</code></pre>",
		},
		{
			name: "without attributes",
			input: "<!-- wp:syntaxhighlighter/code -->\n" +
				"<pre class=\"wp-block-syntaxhighlighter-code\">x = 1</pre>\n" +
				"<!-- /wp:syntaxhighlighter/code -->",
			expected: "<pre><code>x = 1</code></pre>",
		},
		{
			name: "brush aliases are normalized",
			input: "<!-- wp:syntaxhighlighter/code {\"language\":\"JScript\"} -->\n" +
				"<pre class=\"wp-block-syntaxhighlighter-code\">a()</pre>\n" +
				"<!-- /wp:syntaxhighlighter/code -->\n" +
				"<!-- wp:syntaxhighlighter/code {\"language\":\"plain\"} -->\n" +
				"<pre class=\"wp-block-syntaxhighlighter-code\">text</pre>\n" +
				"<!-- /wp:syntaxhighlighter/code -->",
			expected: "<pre><code class=\"language-javascript\">a()</code></pre>\n<pre><code>text</code></pre>",
		},
		{
			name: "with an additional CSS class",
			input: "<!-- wp:syntaxhighlighter/code {\"language\":\"python\",\"className\":\"my-code\"} -->\n" +
				"<pre class=\"wp-block-syntaxhighlighter-code my-code\">pass</pre>\n" +
				"<!-- /wp:syntaxhighlighter/code -->",
			expected: "<pre><code class=\"language-python\">pass</code></pre>",
		},
		{
			name: "language with a space",
			input: "<!-- wp:syntaxhighlighter/code {\"language\":\"Objective C\"} -->\n" +
				"<pre class=\"wp-block-syntaxhighlighter-code\">x</pre>\n" +
				"<!-- /wp:syntaxhighlighter/code -->",
			expected: "<pre><code class=\"language-objective-c\">x</code></pre>",
		},
		{
			name: "with other attributes",
			input: "<!-- wp:syntaxhighlighter/code {\"language\":\"bash\",\"highlightLines\":\"2\"} -->\n" +
				"<pre class=\"wp-block-syntaxhighlighter-code\">echo hi</pre>\n" +
				"<!-- /wp:syntaxhighlighter/code -->",
			expected: "<pre><code class=\"language-bash\">echo hi</code></pre>",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, replaceSyntaxHighlighterCode(tc.input))
		})
	}
}

func TestSyntaxHighlighterCodeInPageContent(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name         string
		htmlInput    string
		expectedBody string
	}{
		{
			name:         "indentation is preserved",
			htmlInput:    "Intro:\n\n[code language=\"python\"]\ndef f(x):\n    if x < 1:\n        return x\n[/code]\n\nAfter.",
			expectedBody: "Intro:\n\n```python\ndef f(x):\n    if x < 1:\n        return x\n```\n\nAfter.",
		},
		{
			name:         "HTML code is kept as code",
			htmlInput:    "[sourcecode lang=\"html\"]\n&lt;div&gt;&amp;&lt;/div&gt;\n[/sourcecode]",
			expectedBody: "```html\n<div>&</div>\n```",
		},
		{
			name:         "escaped shortcode is shown as text",
			htmlInput:    "Example: [[code language=\"php\"]echo 1;[/code]] renders as code.",
			expectedBody: "Example: \\[code language=\"php\"\\]echo 1;\\[/code\\] renders as code.",
		},
		{
			name:         "shortcodes inside the code are not converted",
			htmlInput:    "[code lang=\"php\"]\necho do_shortcode('[audio src=\"a.mp3\"]');\n[/code]",
			expectedBody: "```php\necho do_shortcode('[audio src=\"a.mp3\"]');\n```",
		},
		{
			name: "Gutenberg block",
			htmlInput: "<!-- wp:syntaxhighlighter/code {\"language\":\"python\"} -->\n" +
				"<pre class=\"wp-block-syntaxhighlighter-code\">def f(x):\n    if x &lt; 1:\n        return x</pre>\n" +
				"<!-- /wp:syntaxhighlighter/code -->",
			expectedBody: "```python\ndef f(x):\n    if x < 1:\n        return x\n```",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testMarkdownExtractor(t, tc.htmlInput, tc.expectedBody)
		})
	}
}
