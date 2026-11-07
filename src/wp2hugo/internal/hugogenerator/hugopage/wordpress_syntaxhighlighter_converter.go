package hugopage

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// Support for the "SyntaxHighlighter Evolved" plugin (also used by WordPress.com)
// Ref: https://wordpress.org/plugins/syntaxhighlighter/
//
// Classic editor shortcodes:
//
//	[code language="python"]...[/code]
//	[sourcecode lang="bash"]...[/sourcecode]
//	[python]...[/python]
//
// Gutenberg block:
//
//	<!-- wp:syntaxhighlighter/code {"language":"python"} -->
//	<pre class="wp-block-syntaxhighlighter-code">...</pre>
//	<!-- /wp:syntaxhighlighter/code -->

// Shortcodes that take the language as an attribute
var _syntaxHighlighterGenericShortcodes = []string{"code", "sourcecode"}

// Shortcodes named after the language (brush aliases of SyntaxHighlighter Evolved)
var _syntaxHighlighterLanguageShortcodes = []string{
	"actionscript3", "as3", "bash", "c", "c-sharp", "cf", "coldfusion", "cpp", "csharp", "css",
	"delphi", "diff", "erl", "erlang", "groovy", "html", "java", "javafx", "javascript", "jfx",
	"jscript", "js", "pas", "pascal", "patch", "perl", "php", "pl", "plain", "powershell", "ps",
	"py", "python", "rails", "ror", "ruby", "scala", "shell", "sql", "text", "vb", "vbnet",
	"xhtml", "xml", "xslt",
}

// Brush aliases that Hugo (Chroma) does not know under the same name
var _syntaxHighlighterLanguageAliases = map[string]string{
	"c-sharp": "csharp",
	"erl":     "erlang",
	"jscript": "javascript",
	"patch":   "diff",
	"plain":   "",
	"ps":      "powershell",
	"rails":   "ruby",
	"ror":     "ruby",
	"text":    "",
	"vb":      "vbnet",
	"xhtml":   "html",
}

var (
	_syntaxHighlighterOpeningTag = regexp.MustCompile(`(?i)\[(` +
		joinQuoted(_syntaxHighlighterGenericShortcodes) + "|" +
		joinQuoted(_syntaxHighlighterLanguageShortcodes) +
		`)((?:\s[^\]]*)?)\]`)
	_syntaxHighlighterLanguageAttr = regexp.MustCompile(`(?i)\b(?:language|lang)\s*=\s*["']?([\w#+.-]+)`)

	// Note: the attributes group always participates in the match (possibly empty), since
	// replaceAllStringSubmatchFunc does not support optional groups that don't match
	_syntaxHighlighterGutenbergBlock = regexp.MustCompile(`(?s)<!-- wp:syntaxhighlighter/code\s*((?:\{.*?\})?)\s*-->\s*` +
		`<pre class="wp-block-syntaxhighlighter-code(?:\s[^"]*)?"[^>]*>(.*?)</pre>\s*` +
		`<!-- /wp:syntaxhighlighter/code -->`)
)

func joinQuoted(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, v := range values {
		quoted = append(quoted, regexp.QuoteMeta(v))
	}
	return strings.Join(quoted, "|")
}

// replaceSyntaxHighlighterCode converts SyntaxHighlighter shortcodes and blocks into
// <pre><code class="language-xyz"> elements, which the Markdown converter turns into fenced code blocks.
// Without this, the shortcodes are kept as escaped text and the code's indentation is collapsed.
func replaceSyntaxHighlighterCode(htmlData string) string {
	htmlData = replaceSyntaxHighlighterGutenbergBlocks(htmlData)
	return replaceSyntaxHighlighterShortcodes(htmlData)
}

func replaceSyntaxHighlighterGutenbergBlocks(htmlData string) string {
	return replaceAllStringSubmatchFunc(_syntaxHighlighterGutenbergBlock, htmlData, func(groups []string) string {
		language := ""
		if groups[1] != "" {
			var attributes struct {
				Language string `json:"language"`
			}
			if err := json.Unmarshal([]byte(groups[1]), &attributes); err != nil {
				log.Warn().
					Err(err).
					Str("attributes", groups[1]).
					Msg("Failed to parse SyntaxHighlighter block attributes")
			} else {
				language = normalizeSyntaxHighlighterLanguage(attributes.Language)
			}
		}
		// The block content is already HTML-escaped
		return codeBlockHTML(language, groups[2])
	})
}

func replaceSyntaxHighlighterShortcodes(htmlData string) string {
	var sb strings.Builder
	for {
		loc := _syntaxHighlighterOpeningTag.FindStringSubmatchIndex(htmlData)
		if loc == nil {
			break
		}
		tagName := htmlData[loc[2]:loc[3]]
		attributes := htmlData[loc[4]:loc[5]]

		// The closing tag must use the same name as the opening tag
		closingTag := regexp.MustCompile(`(?i)\[/` + regexp.QuoteMeta(tagName) + `\]`)
		closingLoc := closingTag.FindStringIndex(htmlData[loc[1]:])

		// WordPress displays a shortcode wrapped in an extra pair of brackets as text instead of running it,
		// e.g. "[[code]]" or "[[code]...[/code]]". Keep it as text without the extra brackets.
		// Ref: "Allow [[foo]] syntax for escaping a tag" in https://developer.wordpress.org/reference/functions/do_shortcode_tag/
		if loc[0] > 0 && htmlData[loc[0]-1] == '[' {
			escapedEnd := -1
			if loc[1] < len(htmlData) && htmlData[loc[1]] == ']' {
				escapedEnd = loc[1]
			} else if closingLoc != nil {
				closingEnd := loc[1] + closingLoc[1]
				if closingEnd < len(htmlData) && htmlData[closingEnd] == ']' {
					escapedEnd = closingEnd
				}
			}
			if escapedEnd >= 0 {
				sb.WriteString(htmlData[:loc[0]-1])
				sb.WriteString(htmlData[loc[0]:escapedEnd])
				htmlData = htmlData[escapedEnd+1:]
				continue
			}
		}

		if closingLoc == nil {
			// Not a SyntaxHighlighter shortcode (e.g. "[c]" in regular text), keep it as is
			sb.WriteString(htmlData[:loc[1]])
			htmlData = htmlData[loc[1]:]
			continue
		}

		code := htmlData[loc[1] : loc[1]+closingLoc[0]]
		sb.WriteString(htmlData[:loc[0]])
		sb.WriteString(codeBlockHTML(getSyntaxHighlighterLanguage(tagName, attributes), html.EscapeString(normalizeShortcodeCode(code))))
		htmlData = htmlData[loc[1]+closingLoc[1]:]
	}
	sb.WriteString(htmlData)
	return sb.String()
}

func getSyntaxHighlighterLanguage(tagName string, attributes string) string {
	language := tagName
	if strings.EqualFold(language, "code") || strings.EqualFold(language, "sourcecode") {
		language = ""
		if match := _syntaxHighlighterLanguageAttr.FindStringSubmatch(attributes); match != nil {
			language = match[1]
		}
	}
	return normalizeSyntaxHighlighterLanguage(language)
}

// normalizeSyntaxHighlighterLanguage maps a SyntaxHighlighter brush name to the language name used by Hugo
func normalizeSyntaxHighlighterLanguage(language string) string {
	// A code fence's language can't contain spaces, e.g. "Objective C" -> "objective-c"
	language = strings.ToLower(strings.Join(strings.Fields(language), "-"))
	if alias, ok := _syntaxHighlighterLanguageAliases[language]; ok {
		return alias
	}
	return language
}

// normalizeShortcodeCode returns the raw code inside a shortcode.
// Code written in the visual editor is stored HTML-escaped (e.g. "&lt;div&gt;") while code written
// in the text editor is stored as is, so it is unescaped once here and escaped again by the caller.
func normalizeShortcodeCode(code string) string {
	code = html.UnescapeString(code)
	code = strings.ReplaceAll(code, "\r\n", "\n")
	// The code usually starts on the line after the opening tag
	code = strings.TrimPrefix(code, "\n")
	return strings.TrimRight(code, " \t\n")
}

// _shortcodeBracketEscaper escapes square brackets in the code, so that the converters running after this one
// (e.g. for [audio] or [gallery]) don't treat shortcodes shown in a code sample as real ones.
// The Markdown converter turns the entities back into brackets.
var _shortcodeBracketEscaper = strings.NewReplacer("[", "&#91;", "]", "&#93;")

func codeBlockHTML(language string, escapedCode string) string {
	escapedCode = _shortcodeBracketEscaper.Replace(escapedCode)
	if language == "" {
		return fmt.Sprintf("<pre><code>%s</code></pre>", escapedCode)
	}
	return fmt.Sprintf(`<pre><code class="language-%s">%s</code></pre>`, html.EscapeString(language), escapedCode)
}
