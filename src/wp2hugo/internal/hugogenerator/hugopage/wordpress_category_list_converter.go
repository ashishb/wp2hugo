package hugopage

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/rs/zerolog/log"
)

// Example: "[catlist name="programming" catlink=yes date=yes date\_tag=p excerpt=no numberposts=5 date=no thumbnail=no]"
var _CatlistRegEx = regexp.MustCompile(`\\\[catlist name="([^"]+)" catlink=(yes|no) .* numberposts=([0-9]+).*]`)

// Converts the catlist shortcode to Hugo shortcode using our custom
// shortcode _selectedPostsShortCode
func replaceCatlistWithShortcode(markdownData string) string {
	log.Debug().
		Msg("Replacing catlist with shortcode")

	markdownData = replaceAllStringSubmatchFunc(_CatlistRegEx, markdownData, func(groups []string) string {
		return fmt.Sprintf("{{< catlist category=\"%s\" catlink=%s count=%s >}}",
			wpparser.NormalizeCategoryName(groups[1]), // Normalize the category name the same way as in rest of the code
			groups[2], groups[3])
	})

	return markdownData
}

// Ref: https://gist.github.com/elliotchance/d419395aa776d632d897
func replaceAllStringSubmatchFunc(re *regexp.Regexp, str string, repl func([]string) string) string {
	result := ""
	lastIndex := 0

	var resultSb strings.Builder
	for _, v := range re.FindAllStringSubmatchIndex(str, -1) {
		var groups []string
		for i := 0; i < len(v); i += 2 {
			groups = append(groups, str[v[i]:v[i+1]])
		}

		resultSb.WriteString(str[lastIndex:v[0]] + repl(groups))
		lastIndex = v[1]
	}
	result += resultSb.String()

	return result + str[lastIndex:]
}
