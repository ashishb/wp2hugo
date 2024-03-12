package hugogenerator

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"regexp"
	"strings"
)

// Example:  [caption id="" align="aligncenter" width="599"]<img class=""
//
//	src="http://static.cdn-seekingalpha.com/uploads/2011/8/7/saupload_us_tax_rates_2.png"
//	alt="Dividend tax rate in the US" width="599" height="283" />
//	Dividend tax rate in the US
//	[/caption]
//
// the important fields to extract are "align", "width", "src", "alt"
var _CaptionRegEx1 = regexp.MustCompile(`\[caption [^ ]* align="([^"]+)" width="([^"]+)"\]` +
	`.*?` +
	`<img class="[^"]*?" src="([^"]+)" alt="([^"]*?)" width="([^"]*?)" height="([^"]*?)" />` +
	`.+?` +
	`\[/caption\]`)

// No alt
var _CaptionRegEx2 = regexp.MustCompile(`\[caption [^ ]* align="([^"]+)" width="([^"]+)"\]` +
	`.*?` +
	`<img class="[^"]*?" src="([^"]+)" width="([^"]*?)" height="([^"]*?)" />` +
	`.+?` +
	`\[/caption\]`)

// Converts the WordPress's caption shortcode to Hugo shortcode "figure"
// https://adityatelange.github.io/hugo-PaperMod/posts/papermod/papermod-faq/#centering-image-in-markdown
func replaceCaptionWithFigure(htmlData string) string {
	log.Debug().
		Msg("Replacing caption with figure")

	htmlData = replaceAllStringSubmatchFunc(_CaptionRegEx1, htmlData, captionReplacementFunction)
	htmlData = replaceAllStringSubmatchFunc(_CaptionRegEx2, htmlData, captionReplacementFunction)
	return htmlData
}

func captionReplacementFunction(groups []string) string {
	const replacementQuote = "'"
	src := groups[3]
	// These character creates problem in Hugo's markdown
	src = strings.ReplaceAll(src, " ", "%20")
	src = strings.ReplaceAll(src, "_", "%5F")
	alt := ""
	if len(groups) > 4 {
		alt := groups[4]
		for _, s := range []string{"\"", "“", "”", "&quot;"} {
			alt = strings.ReplaceAll(alt, s, replacementQuote)
		}
	}
	return fmt.Sprintf(`{{< figure align=%s width=%s src="%s" alt="%s" >}}`,
		groups[1], groups[2], src, alt)
}
