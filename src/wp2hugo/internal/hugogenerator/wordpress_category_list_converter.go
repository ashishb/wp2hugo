package hugogenerator

import (
	"fmt"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/rs/zerolog/log"
	"regexp"
)

// Example: "[catlist name="programming" catlink=yes date=yes date\_tag=p excerpt=no numberposts=5 date=no thumbnail=no]"
var _CatlistRegEx = regexp.MustCompile(`\\\[catlist name="([^"]+)" catlink=(yes|no) .* numberposts=([0-9]+).*]`)

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

	for _, v := range re.FindAllSubmatchIndex([]byte(str), -1) {
		groups := []string{}
		for i := 0; i < len(v); i += 2 {
			groups = append(groups, str[v[i]:v[i+1]])
		}

		result += str[lastIndex:v[0]] + repl(groups)
		lastIndex = v[1]
	}

	return result + str[lastIndex:]
}
