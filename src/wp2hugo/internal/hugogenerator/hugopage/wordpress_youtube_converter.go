package hugopage

import (
	"fmt"
	"regexp"

	"github.com/rs/zerolog/log"
)

// Example: Plain-text Youtube URLs on their own line in post content are turned by WP into embeds
// The YouTube Lyte plug-in additionally uses "httpa://" for audio and "httpv://" for video embeds:
// https://wordpress.com/plugins/wp-youtube-lyte
var _YoutubeRegEx = regexp.MustCompile(`(?m)(^|\s)(?:\[embed\])?http[sav]?://(?:m\.|www\.)?(?:youtu\.be|youtube\.com)/(?:watch|w)\?v=([^&\s\[\]]+)(?:\[\/embed\])?`)

// Gutenberg Youtube embed into figure:
// <!-- wp:embed {"url":"https://www.youtube.com/watch?v=7l6FjphZXsk","type":"video","providerNameSlug":"youtube","responsive":true,"align":"full","className":"wp-embed-aspect-16-9 wp-has-aspect-ratio"} -->
// <figure class="wp-block-embed alignfull is-type-video is-provider-youtube wp-block-embed-youtube wp-embed-aspect-16-9 wp-has-aspect-ratio"><div class="wp-block-embed__wrapper">
// https://www.youtube.com/watch?v=7l6FjphZXsk
// </div></figure>
// <!-- /wp:embed -->
var _YoutubeGutenbergRegEx = regexp.MustCompile(`(?ms)(^|\s)<!-- wp:embed {"url":"[^"]+v=([^"]+)".*?<!-- /wp:embed -->`)

func replacePlaintextYoutubeURL(htmlData string) string {
	log.Debug().
		Msg("Replacing Youtube URLs with embeds")

	htmlData = replaceAllStringSubmatchFunc(_YoutubeGutenbergRegEx, htmlData, YoutubeReplacementFunction)
	htmlData = replaceAllStringSubmatchFunc(_YoutubeRegEx, htmlData, YoutubeReplacementFunction)
	return htmlData
}

func YoutubeReplacementFunction(groups []string) string {
	return fmt.Sprintf(`%s{{< youtube %s >}}`, groups[1], groups[2])
}
