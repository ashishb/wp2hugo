package hugopage

import (
	"fmt"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
	"regexp"
)

var (
	// Match consecutive <br>
	_consecutiveBRRegex = regexp.MustCompile(`(?mi)(<br\s*/?>\s*){2,}`)
)

const (
	_consecutiveBrCustomTag = "consecutive-br"
	// "  \n" gets trimmed to "\n" by the markdown converter :/
	_doubleSpaceWithNewline = "{{ double-space-with-newline }}"
)

func convertConsecutiveBRToCustomTag(htmlContent string) string {
	output := _consecutiveBRRegex.ReplaceAllString(htmlContent, "<"+_consecutiveBrCustomTag+" />")
	if output != htmlContent {
		log.Debug().
			Str("input", htmlContent).
			Str("output", output).
			Msg("Converting consecutive <br> to custom tag")
	}
	return output
}

func convertCustomBRToNewline() md.Plugin {
	return func(c *md.Converter) []md.Rule {
		return []md.Rule{
			{
				Filter: []string{_consecutiveBrCustomTag},
				Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
					log.Debug().
						Str("content", content).
						Str("text", selec.Text()).
						Msg("Converting custom BR to newline")
					text := fmt.Sprintf("\n\n%s", content)
					return &text
				},
			},
		}
	}
}

// Replace <br> with "  \n"
// This works as long as there are no consecutive <br>
func convertBrToNewline() md.Plugin {
	return func(c *md.Converter) []md.Rule {
		return []md.Rule{
			{
				Filter: []string{"br"},
				Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
					// Ref: https://github.com/ashishb/wp2hugo/issues/12
					log.Info().
						Str("content", content).
						Msg("Converting <br> to newline")
					text := _doubleSpaceWithNewline
					return &text
				},
			},
		}
	}
}
