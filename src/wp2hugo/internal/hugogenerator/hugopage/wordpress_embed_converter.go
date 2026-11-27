package hugopage

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// WordPress "[embed]URL[/embed]" shortcode, optionally with attributes, e.g. "[embed width="500"]URL[/embed]".
// This runs on Markdown, where the HTML-to-Markdown converter escapes the brackets as "\[embed\]".
var _embedShortcodeRegEx = regexp.MustCompile(`\\?\[embed(?:\s[^\]]*?)?\\?\](.*?)\\?\[/embed\\?\]`)

// WordPress turns a URL that is alone on its own line into an embed (oEmbed).
// Gutenberg embed blocks also end up as the bare URL on its own line after the HTML-to-Markdown conversion:
// <!-- wp:embed {"url":"https://vimeo.com/76979871","type":"video","providerNameSlug":"vimeo"} -->
// <figure class="wp-block-embed is-type-video is-provider-vimeo wp-block-embed-vimeo"><div class="wp-block-embed__wrapper">
// https://vimeo.com/76979871
// </div></figure>
// <!-- /wp:embed -->
// URLs inside a sentence are links, not embeds, and are left untouched.
var (
	// Only public videos (vimeo.com/ID). Unlisted videos (vimeo.com/ID/HASH) need the hash,
	// which Hugo's "vimeo" shortcode does not support, so they are kept as links.
	_vimeoURLOnOwnLineRegEx = regexp.MustCompile(`(?m)^[ \t]*https?://(?:www\.)?vimeo\.com/(\d+)/?(?:[?#]\S*)?[ \t]*$`)
	// Note: "_" in the post ID may have been escaped as "\_" in Markdown.
	_instagramURLOnOwnLineRegEx = regexp.MustCompile(`(?m)^[ \t]*https?://(?:www\.)?instagram\.com/p/((?:[A-Za-z0-9-]|\\?_)+)/?(?:\?\S*)?[ \t]*$`)
)

// unwrapEmbedShortcodes replaces "[embed]URL[/embed]" with the URL, so that the URL can be converted
// to a Hugo shortcode by the provider-specific steps, or stays as a plain link otherwise.
func unwrapEmbedShortcodes(markdown string) string {
	return _embedShortcodeRegEx.ReplaceAllString(markdown, "$1")
}

func replacePlaintextVimeoAndInstagramURLs(markdown string) string {
	log.Debug().
		Msg("Replacing Vimeo and Instagram URLs with embeds")
	markdown = replaceAllStringSubmatchFunc(_vimeoURLOnOwnLineRegEx, markdown, func(groups []string) string {
		return fmt.Sprintf("{{< vimeo %s >}}", groups[1])
	})
	markdown = replaceAllStringSubmatchFunc(_instagramURLOnOwnLineRegEx, markdown, func(groups []string) string {
		return fmt.Sprintf("{{< instagram %s >}}", strings.ReplaceAll(groups[1], `\_`, "_"))
	})
	return markdown
}
