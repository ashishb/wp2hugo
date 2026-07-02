package hugopage

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// Examples:
//  1. [video mp4="/wp-content/uploads/2026/05/video.mp4"][/video]
//  2. [video src="video.mp4"]
//  3. [video mp4="source.mp4" webm="source.webm"]
//  4. Gutenberg editor directly writes video HTML, like:
//     <figure class="wp-block-video"><video controls src="/wp-content/uploads/2026/05/video.mp4"></video></figure>
//     Gutenberg can optionally nest <figcaption> into <figure>, below <video>. We disregard it here.
//
// Reference: https://wordpress.org/documentation/article/video-shortcode/
var (
	_VideoShortCodeRegEx = regexp.MustCompile(`\[video ([^\]]+)\](?:.*)(?:\[\/video\])?`)
	_VideoHTMLRegEx      = regexp.MustCompile(`<figure (?:.*?)class="(?:.*?)wp-block-video(?:.*?)">\s*<video ([^<>]*?)\/?>(?:<\/video>)?(?:[\s\S]*?)</figure>`)
)

var (
	_mp4RegEx  = regexp.MustCompile(`mp4="([^"]+)"`)
	_m4vRegEx  = regexp.MustCompile(`m4v="([^"]+)"`)
	_webmRegEx = regexp.MustCompile(`webm="([^"]+)"`)
	_ogvRegEx  = regexp.MustCompile(`ogv="([^"]+)"`)
)

func replaceVideoShortCode(htmlData string) string {
	log.Debug().
		Msg("Replacing Video shortcodes")
	htmlData = replaceAllStringSubmatchFunc(_VideoShortCodeRegEx, htmlData, VideoReplacementFunction)
	htmlData = replaceAllStringSubmatchFunc(_VideoHTMLRegEx, htmlData, VideoReplacementFunction)
	return htmlData
}

func printVideoShortCode(src string) string {
	// These characters create problems in Hugo's markdown
	src = strings.ReplaceAll(src, " ", "%20")
	src = strings.ReplaceAll(src, "_", "%5F")
	return fmt.Sprintf(`{{< video src="%s" >}}`, src)
}

func VideoReplacementFunction(groups []string) string {
	args := groups[1]

	// We look for, in this order: mp4, m4v, src, webm, ogv; then take the first.
	// mp4 and m4v are the most widely supported video formats.
	// Reference: https://developer.mozilla.org/en-US/docs/Web/Media/Formats/Video_codecs

	mp4 := _mp4RegEx.FindStringSubmatch(args)
	if mp4 != nil {
		return printVideoShortCode(mp4[1])
	}

	m4v := _m4vRegEx.FindStringSubmatch(args)
	if m4v != nil {
		return printVideoShortCode(m4v[1])
	}

	src := _srcRegEx.FindStringSubmatch(args)
	if src != nil {
		return printVideoShortCode(src[1])
	}

	webm := _webmRegEx.FindStringSubmatch(args)
	if webm != nil {
		return printVideoShortCode(webm[1])
	}

	ogv := _ogvRegEx.FindStringSubmatch(args)
	if ogv != nil {
		return printVideoShortCode(ogv[1])
	}

	return ""
}
