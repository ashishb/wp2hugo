package hugopage

import (
	"fmt"
	"regexp"

	"github.com/rs/zerolog/log"
)

// Example: Plain-text Youtube URLs on their own line in post content are turned by WP into embeds
var _YoutubeRegEx = regexp.MustCompile(`(?m)(^|\s)https?://(?:m\.|www\.)?(?:youtu\.be|youtube\.com)/(?:watch|w)\?v=([^&\s]+)`)

func replacePlaintextYoutubeURL(htmlData string) string {
	log.Debug().
		Msg("Replacing Youtube URLs with embeds")
	return replaceAllStringSubmatchFunc(_YoutubeRegEx, htmlData, YoutubeReplacementFunction)
}

func YoutubeReplacementFunction(groups []string) string {
	return fmt.Sprintf(`%s{{< youtube %s >}}`, groups[1], groups[2])
}
