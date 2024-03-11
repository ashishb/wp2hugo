package hugogenerator

import (
	"github.com/rs/zerolog/log"
	"regexp"
)

// Example: "[catlist name="programming" catlink=yes date=yes date\_tag=p excerpt=no numberposts=5 date=no thumbnail=no]"
var _CatlistRegEx = regexp.MustCompile(`\\\[catlist name="([^"]+)" catlink=(yes|no) .* numberposts=([0-9]+).*]`)

func replaceCatlistWithShortcode(markdownData string) string {
	log.Debug().
		Msg("Replacing catlist with shortcode")
	markdownData = _CatlistRegEx.ReplaceAllString(markdownData, "{{< catlist category=\"$1\" catlink=$2 count=$3 >}}")
	return markdownData
}
