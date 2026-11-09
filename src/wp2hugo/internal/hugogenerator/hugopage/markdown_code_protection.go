package hugopage

import (
	"fmt"
	"regexp"
	"strings"
)

// The Markdown post-processing steps (absolute to relative links, YouTube URLs to shortcodes,
// collapsing blank lines, ...) are meant for the text of a post, not for code samples.
// protectCode replaces fenced code blocks and inline code spans with placeholders before
// those steps run, and restoreCode puts the original code back afterwards.

// NUL can't appear in the content, since it is not a valid XML character in the WordPress export
const _codePlaceholderFormat = "\x00wp2hugo-code-%d\x00"

var _fencedCodeBlockOpening = regexp.MustCompile("^([ \t]*)(`{3,}|~{3,})")

// The HTML to Markdown converter collapses consecutive blank lines in the whole output, including
// code blocks (e.g. the two blank lines between Python functions become one).
// Blank lines inside <pre> are replaced with this marker before the conversion and restored right after it.
const _blankLineInCodeMarker = "{{ wp2hugo-blank-line-in-code }}"

var _preElement = regexp.MustCompile(`(?is)(<pre\b[^>]*>)(.*?)(</pre>)`)

func protectBlankLinesInPre(htmlContent string) string {
	return replaceAllStringSubmatchFunc(_preElement, htmlContent, func(groups []string) string {
		lines := strings.Split(groups[2], "\n")
		firstCodeLine, lastCodeLine := -1, -1
		for i, line := range lines {
			if strings.TrimSpace(line) != "" {
				if firstCodeLine < 0 {
					firstCodeLine = i
				}
				lastCodeLine = i
			}
		}
		// Only the blank lines between two lines of code; leading and trailing blank lines are left
		// to the converter as before
		for i := firstCodeLine + 1; i < lastCodeLine; i++ {
			if strings.TrimSpace(lines[i]) == "" {
				lines[i] = _blankLineInCodeMarker
			}
		}
		return groups[1] + strings.Join(lines, "\n") + groups[3]
	})
}

func restoreBlankLinesInCode(markdown string) string {
	return strings.ReplaceAll(markdown, _blankLineInCodeMarker, "")
}

type protectedCode []string

func protectCode(markdown string) (string, protectedCode) {
	var code protectedCode
	markdown = protectFencedCodeBlocks(markdown, &code)
	markdown = protectInlineCode(markdown, &code)
	return markdown, code
}

func restoreCode(markdown string, code protectedCode) string {
	// Restore in reverse order: inline code spans are protected after the fenced code blocks,
	// so the text of a span may contain the placeholder of a fenced code block, but not the other way around
	for i := len(code) - 1; i >= 0; i-- {
		markdown = strings.Replace(markdown, fmt.Sprintf(_codePlaceholderFormat, i), code[i], 1)
	}
	return markdown
}

func (c *protectedCode) add(code string) string {
	*c = append(*c, code)
	return fmt.Sprintf(_codePlaceholderFormat, len(*c)-1)
}

// protectFencedCodeBlocks replaces each fenced code block, from its opening to its closing fence,
// with a placeholder. The indentation of the opening fence is kept, so that a code block inside
// a list item stays inside that list item.
func protectFencedCodeBlocks(markdown string, code *protectedCode) string {
	lines := strings.Split(markdown, "\n")
	result := make([]string, 0, len(lines))
	for i := 0; i < len(lines); i++ {
		match := _fencedCodeBlockOpening.FindStringSubmatch(lines[i])
		if match == nil {
			result = append(result, lines[i])
			continue
		}

		indentation, fence := match[1], match[2]
		end := len(lines) - 1 // An unclosed code block runs until the end of the document
		for j := i + 1; j < len(lines); j++ {
			if isClosingFence(lines[j], fence) {
				end = j
				break
			}
		}

		block := strings.Join(lines[i:end+1], "\n")
		result = append(result, indentation+code.add(strings.TrimPrefix(block, indentation)))
		i = end
	}
	return strings.Join(result, "\n")
}

// isClosingFence reports whether the line closes a code block opened with the given fence:
// the same character, at least as many times, and nothing else on the line.
func isClosingFence(line string, openingFence string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < len(openingFence) {
		return false
	}
	return strings.Trim(trimmed, openingFence[:1]) == ""
}

// protectInlineCode replaces inline code spans (e.g. `curl https://example.com`) with placeholders.
// A code span starts with a run of backticks and ends with a run of the same length,
// within the same paragraph.
func protectInlineCode(markdown string, code *protectedCode) string {
	var sb strings.Builder
	for i := 0; i < len(markdown); {
		if markdown[i] != '`' {
			sb.WriteByte(markdown[i])
			i++
			continue
		}

		openingLength := backtickRunLength(markdown, i)
		if i > 0 && markdown[i-1] == '\\' {
			// An escaped backtick ("\`") does not start a code span
			sb.WriteString(markdown[i : i+openingLength])
			i += openingLength
			continue
		}

		end := findClosingBacktickRun(markdown, i+openingLength, openingLength)
		if end < 0 {
			sb.WriteString(markdown[i : i+openingLength])
			i += openingLength
			continue
		}

		sb.WriteString(code.add(markdown[i:end]))
		i = end
	}
	return sb.String()
}

// findClosingBacktickRun returns the index right after a run of exactly n backticks
// starting at or after start, or -1 if the paragraph ends before such a run.
func findClosingBacktickRun(markdown string, start int, n int) int {
	for j := start; j < len(markdown); {
		if strings.HasPrefix(markdown[j:], "\n\n") {
			return -1
		}
		if markdown[j] != '`' {
			j++
			continue
		}
		runLength := backtickRunLength(markdown, j)
		if runLength == n {
			return j + runLength
		}
		j += runLength
	}
	return -1
}

func backtickRunLength(s string, start int) int {
	n := 0
	for start+n < len(s) && s[start+n] == '`' {
		n++
	}
	return n
}
