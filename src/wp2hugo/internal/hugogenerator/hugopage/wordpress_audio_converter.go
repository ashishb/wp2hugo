package hugopage

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// Examples:
//  1. [audio mp3="/wp-content/uploads/sites/3/2020/07/session_2020-07-02.mp3"][/audio]
//  2. [audio src="audio-source.mp3"]
//  3. [audio mp3="source.mp3" ogg="source.ogg" wav="source.wav" m4a="source.m4a"]
//  4. [audio] is allowed by WP but is not covered here since WP extracts the first link to mp3/ogg/wav/m4a found in post.
//     this case is inconvenient for us and pretty niche.
//  5. Gutenberg editor directly writes audio HTML, like :
//     <figure class="wp-block-audio"><audio src="/wp-content/uploads/sites/3/2020/07/session_2020-07-02.mp3" controls="controls"></audio></figure>
//     Gutenberg can optionnaly nest <figcaption> into <figure>, below <audio>. We disregard it here.
//
// Reference : https://wordpress.org/documentation/article/audio-shortcode/
var (
	_AudioShortCodeRegEx = regexp.MustCompile(`\[audio ([^\]]+)\](?:.*)(?:\[\/audio\])?`)
	_AudioHTMLRegEx      = regexp.MustCompile(`<figure (?:.*?)class="(?:.*?)wp-block-audio(?:.*?)">\s*<audio ([^<>]*?)\/?>(?:<\/audio>)?(?:[\s\S]*?)</figure>`)
)

var (
	_srcRegEx = regexp.MustCompile(`src="([^"]+)"`)
	_mp3RegEx = regexp.MustCompile(`mp3="([^"]+)"`)
	_m4aRegEx = regexp.MustCompile(`m4a="([^"]+)"`)
	_oggRegEx = regexp.MustCompile(`ogg="([^"]+)"`)
	_wavRegEx = regexp.MustCompile(`wav="([^"]+)"`)
)

func replaceAudioShortCode(htmlData string) string {
	log.Debug().
		Msg("Replacing Audio shortcodes")
	htmlData = replaceAllStringSubmatchFunc(_AudioShortCodeRegEx, htmlData, AudioReplacementFunction)
	htmlData = replaceAllStringSubmatchFunc(_AudioHTMLRegEx, htmlData, AudioReplacementFunction)
	return htmlData
}

func printAudioShortCode(src string) string {
	// These characters create problems in Hugo's markdown
	src = strings.ReplaceAll(src, " ", "%20")
	src = strings.ReplaceAll(src, "_", "%5F")
	return fmt.Sprintf(`{{< audio src="%s" >}}`, src)
}

func AudioReplacementFunction(groups []string) string {
	args := groups[1]

	// We look for, in this order : m4a, mp3, src, ogg, wav ; then take the first.
	// Reason is m4a and mp3 are the most widely supported.
	// We could support all alternatives at once, using <source>, and let the browser decide
	// but is it worth the trouble ?
	// Reference : https://developer.mozilla.org/en-US/docs/Web/Media/Formats/Audio_codecs

	m4a := _m4aRegEx.FindStringSubmatch(args)
	if m4a != nil {
		return printAudioShortCode(m4a[1])
	}

	mp3 := _mp3RegEx.FindStringSubmatch(args)
	if mp3 != nil {
		return printAudioShortCode(mp3[1])
	}

	src := _srcRegEx.FindStringSubmatch(args)
	if src != nil {
		return printAudioShortCode(src[1])
	}

	ogg := _oggRegEx.FindStringSubmatch(args)
	if ogg != nil {
		return printAudioShortCode(ogg[1])
	}

	wav := _wavRegEx.FindStringSubmatch(args)
	if wav != nil {
		return printAudioShortCode(wav[1])
	}

	return ""
}
