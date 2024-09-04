package hugopage

import (
	"fmt"
	"regexp"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
)

var youtubeID = regexp.MustCompile(`youtube\.com/embed/([^\&\?\/]+)`)
var googleMapsID = regexp.MustCompile(`google\.com/maps/d/.*embed\?mid=([0-9A-Za-z-_]+)`)
var gistUrl = regexp.MustCompile(`gist\.github\.com/([^/]+)/([0-9a-f]+)`)
var gistMarkdown = regexp.MustCompile(`\\\[gist .*\]`)

func getMarkdownConverter() *md.Converter {
	converter := md.NewConverter("", true, nil)
	converter.Use(getYouTubeForHugoConverter())
	converter.Use(getGoogleMapsEmbedForHugoConverter())
	converter.Use(convertCustomBRToNewline())
	converter.Use(convertBrToNewline())
	converter.Use(convertGistURLsToShortcodes())
	converter.Use(handleBlockEditorImages())
	return converter
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

// <figure class="wp-block-image size-large"><a href="https://blog.gripdev.xyz/wp-content/uploads/2024/03/image.png"><img src="https://blog.gripdev.xyz/wp-content/uploads/2024/03/image.png?w=1024" alt="" class="wp-image-1663" /></a></figure>
func handleBlockEditorImages() md.Plugin {
	return func(c *md.Converter) []md.Rule {
		return []md.Rule{
			{
				Filter: []string{"figure"},
				Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
					classes := selec.AttrOr("class", "")
					if !strings.Contains(classes, "wp-block-image") {
						return &content
					}
					imgUrl := selec.Find("figure.wp-block-image a img").First()
					imgSrc := imgUrl.AttrOr("src", "")
					if imgSrc == "" {
						return &content
					}
					strippedImgSrc := strings.Split(imgSrc, "?")[0]
					text := strings.Replace(content, imgSrc, strippedImgSrc, 1)
					log.Debug().
						Str("imgSrc", imgSrc).
						Msg("Block editor image found, stripping params")
					return &text
				},
			},
		}
	}
}
