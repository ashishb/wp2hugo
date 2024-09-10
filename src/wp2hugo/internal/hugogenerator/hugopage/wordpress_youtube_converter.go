package hugopage

import (
	"fmt"
	"regexp"

	"github.com/rs/zerolog/log"
)

// Example: Plain-text Youtube URLs on their own line in post content are turned by WP into embeds
var _YoutubeRegEx = regexp.MustCompile(`(?m)(?:^|\s)https?://(?:m.|www.)?(?:youtu.be|youtube.com)/(?:watch|w)\?v=([^&\s]+)`)

func replaceYoutubeURL(htmlData string) string {
	log.Debug().
		Msg("Replacing Youtube URLs with embeds")

	htmlData = replaceAllStringSubmatchFunc(_YoutubeRegEx, htmlData, YoutubeReplacementFunction)

	return htmlData
}

func YoutubeReplacementFunction(groups []string) string {
	return fmt.Sprintf(`{{< youtube %s >}}`, groups[1])
}
