package hugopage

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
)

var (
	youtubeID    = regexp.MustCompile(`youtube\.com/embed/([^\&\?\/]+)`)
	googleMapsID = regexp.MustCompile(`google\.com/maps/d/.*embed\?mid=([0-9A-Za-z-_]+)`)
	vimeoID      = regexp.MustCompile(`player\.vimeo\.com/video/(\d+)`)
	gistUrl      = regexp.MustCompile(`gist\.github\.com/([^/]+)/([0-9a-f]+)`)
	gistMarkdown = regexp.MustCompile(`\\\[gist .*\]`)
)

func getMarkdownConverter() *md.Converter {
	converter := md.NewConverter("", true, nil)
	// Convert <table> into a Markdown (GFM) table, otherwise the cells are concatenated into a single line of text.
	// Hugo supports GFM tables by default.
	// Note: tables without a header row get an empty header row, since Markdown tables require one.
	converter.Use(plugin.Table())
	converter.Use(escapePipesInTableCellCode())
	// Rules for the same tag are tried in reverse order of registration, so this fallback must be registered first
	converter.Use(keepOtherIframesAsHTML())
	converter.Use(getYouTubeForHugoConverter())
	converter.Use(getGoogleMapsEmbedForHugoConverter())
	converter.Use(getVimeoForHugoConverter())
	converter.Use(convertCustomBRToNewline())
	converter.Use(convertBrToNewline())
	converter.Use(convertGistURLsToShortcodes())
	return converter
}

// escapePipesInTableCellCode handles inline code inside a table cell. In a GFM table, an unescaped "|"
// ends the cell even inside a code span, so it has to be written as "\|".
// Ref: "Include a pipe in a cell's content by escaping it, including inside other inline spans"
// https://github.github.com/gfm/#example-200
// Other <code> elements are left to the default rule.
func escapePipesInTableCellCode() md.Plugin {
	return func(c *md.Converter) []md.Rule {
		return []md.Rule{
			{
				Filter: []string{"code"},
				Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
					if selec.ParentsFiltered("td, th").Length() == 0 || selec.ParentsFiltered("pre").Length() > 0 {
						return nil
					}

					code := strings.ReplaceAll(selec.Text(), "|", `\|`)
					fence := strings.Repeat("`", longestBacktickRun(code)+1)
					// A code span can't start or end with a backtick next to the fence
					if strings.HasPrefix(code, "`") || strings.HasSuffix(code, "`") {
						code = " " + code + " "
					}
					text := md.AddSpaceIfNessesary(selec, fence+code+fence)
					return &text
				},
			},
		}
	}
}

func longestBacktickRun(s string) int {
	longest, current := 0, 0
	for _, r := range s {
		if r == '`' {
			current++
			longest = max(longest, current)
		} else {
			current = 0
		}
	}
	return longest
}

// Ref: https://github.com/JohannesKaufmann/html-to-markdown/blob/master/plugin/iframe_youtube.go
// YoutubeEmbed registers a rule (for iframes) and
// returns a Hugo markdown compatible representation
func getYouTubeForHugoConverter() md.Plugin {
	return func(c *md.Converter) []md.Rule {
		return []md.Rule{
			{
				Filter: []string{"iframe"},
				Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
					src := selec.AttrOr("src", "")
					if !strings.Contains(src, "youtube.com") {
						return nil
					}

					parts := youtubeID.FindStringSubmatch(src)
					if len(parts) != 2 {
						return nil
					}
					id := parts[1]
					text := fmt.Sprintf("{{< youtube id=\"%s\" >}}", id)
					log.Debug().
						Str("id", id).
						Msg("Youtube video found")
					return &text
				},
			},
		}
	}
}

func getGoogleMapsEmbedForHugoConverter() md.Plugin {
	return func(c *md.Converter) []md.Rule {
		return []md.Rule{
			{
				Filter: []string{"iframe"},
				Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
					src := selec.AttrOr("src", "")
					width := selec.AttrOr("width", "640")
					height := selec.AttrOr("height", "480")
					parts := googleMapsID.FindStringSubmatch(src)
					if len(parts) != 2 {
						return nil
					}
					id := parts[1]
					log.Debug().
						Str("id", id).
						Msg("Google Maps embed found")
					text := fmt.Sprintf("{{< googlemaps src=\"%s\" width=%s height=%s >}}", id, width, height)
					return &text
				},
			},
		}
	}
}

// Example: <iframe src="https://player.vimeo.com/video/76979871" width="640" height="360" allowfullscreen></iframe>
func getVimeoForHugoConverter() md.Plugin {
	return func(c *md.Converter) []md.Rule {
		return []md.Rule{
			{
				Filter: []string{"iframe"},
				Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
					src := selec.AttrOr("src", "")
					parts := vimeoID.FindStringSubmatch(src)
					if len(parts) != 2 {
						return nil
					}
					// Unlisted videos need the "h" (hash) parameter, which Hugo's "vimeo" shortcode does not support.
					// Such iframes are kept as they are.
					if u, err := url.Parse(src); err != nil || u.Query().Has("h") {
						return nil
					}
					id := parts[1]
					log.Debug().
						Str("id", id).
						Msg("Vimeo video found")
					text := fmt.Sprintf("{{< vimeo %s >}}", id)
					return &text
				},
			},
		}
	}
}

// keepOtherIframesAsHTML keeps iframes that are not converted to a Hugo shortcode (e.g. Dailymotion, SoundCloud,
// unlisted Vimeo videos) as raw HTML. Without this rule, the converter drops them, and the embedded content is lost.
// Hugo renders raw HTML since the generated config enables "markup.goldmark.renderer.unsafe".
func keepOtherIframesAsHTML() md.Plugin {
	return func(c *md.Converter) []md.Rule {
		return []md.Rule{
			{
				Filter: []string{"iframe"},
				Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
					html, err := goquery.OuterHtml(selec)
					if err != nil {
						log.Warn().
							Err(err).
							Msg("Failed to render iframe as HTML")
						return nil
					}
					log.Debug().
						Str("src", selec.AttrOr("src", "")).
						Msg("Keeping iframe as HTML")
					return &html
				},
			},
		}
	}
}

func extractShortcodeFromGistUrl(content string) string {
	parts := gistUrl.FindStringSubmatch(content)
	if len(parts) != 3 {
		return content
	}
	user := parts[1]
	id := parts[2]
	text := fmt.Sprintf("{{< gist %s %s >}}", user, id)
	log.Debug().
		Str("user", user).
		Str("id", id).
		Msg("Gist URL found")
	return text
}

func convertGistURLsToShortcodes() md.Plugin {
	return func(c *md.Converter) []md.Rule {
		return []md.Rule{
			// Handle new embed style from block editor
			{
				Filter: []string{"figure"},
				Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
					classes := selec.AttrOr("class", "")
					if !strings.Contains(classes, "is-provider-embed-handler") {
						return &content
					}
					shortcode := extractShortcodeFromGistUrl(content)
					return &shortcode
				},
			},
			// Handle basic `[gist url] embed`
			{
				Filter: []string{"body"},
				Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
					text := gistMarkdown.ReplaceAllStringFunc(content, func(s string) string {
						return extractShortcodeFromGistUrl(content)
					})
					return &text
				},
			},
		}
	}
}
