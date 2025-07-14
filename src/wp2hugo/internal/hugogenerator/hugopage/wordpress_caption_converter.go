package hugopage

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
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
	`<img.*?src="([^"]+)" alt="([^"]*?)".*?/>` +
	`(?:</a>)?` +
	`(.+?)` +
	`\[/caption\]`)

// No alt
var _CaptionRegEx2 = regexp.MustCompile(`\[caption [^ ]* align="([^"]+)" width="([^"]+)"\]` +
	`.*?` +
	`<img.*?src="([^"]+)".*?/>` +
	`(?:</a>)?` +
	`(.+?)` +
	`\[/caption\]`)

// Gutenberg image blocs, no figcaption :
// <!-- wp:image {"align":"center","id":3875,"sizeSlug":"large","className":"is-style-default"} -->
// <div class="wp-block-image is-style-default"><figure class="aligncenter size-large"><img src="https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2016/03/Shooting-Minh-Ly-0155-_DSC0155-Minh-Ly-WEB-1100x1100.jpg" alt="" class="wp-image-3875"/></figure></div>
// <!-- /wp:image -->
var _FigureRegexNoCaption = regexp.MustCompile(`(?ms)<!-- wp:image.*?-->` +
	`.*?<figure.*?` +
	`<img.*?src="([^"]+)".*?alt="([^"]*)".*?>` +
	`</figure>.*?` +
	`<!-- /wp:image -->`)

// Gutenberg image blocs, with figcaption :
// <!-- wp:image {"align":"center","id":3875,"sizeSlug":"large","className":"is-style-default"} -->
// <div class="wp-block-image is-style-default"><figure class="aligncenter size-large"><img src="https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2016/03/Shooting-Minh-Ly-0155-_DSC0155-Minh-Ly-WEB-1100x1100.jpg" alt="" class="wp-image-3875"/><figcaption>Minh-Ly</figcaption></figure></div>
// <!-- /wp:image -->
var _FigureRegexCaption = regexp.MustCompile(`(?ms)<!-- wp:image.*?-->` +
	`.*?<figure.*?` +
	`<img.*?src="([^"]+)".*?alt="([^"]*)".*?>` +
	`<figcaption>(.*?)</figcaption>` +
	`</figure>.*?` +
	`<!-- /wp:image -->`)

// Converts the WordPress's caption shortcode to Hugo shortcode "figure"
// https://adityatelange.github.io/hugo-PaperMod/posts/papermod/papermod-faq/#centering-image-in-markdown
func replaceCaptionWithFigure(htmlData string) string {
	log.Debug().
		Msg("Replacing caption with figure")

	htmlData = replaceAllStringSubmatchFunc(_CaptionRegEx1, htmlData, captionReplacementFunction)
	htmlData = replaceAllStringSubmatchFunc(_CaptionRegEx2, htmlData, captionReplacementFunction)
	return htmlData
}

func replaceImageBlockWithFigure(htmlData string) string {
	log.Debug().
		Msg("Replacing Gutenberg image with figure")

	htmlData = replaceAllStringSubmatchFunc(_FigureRegexNoCaption, htmlData, imageBlockReplacementFunction)
	htmlData = replaceAllStringSubmatchFunc(_FigureRegexCaption, htmlData, imageBlockReplacementFunction)
	return htmlData
}

func sanitizeLinks(src string) string {
	// These character creates problem in Hugo's markdown
	src = strings.ReplaceAll(src, " ", "%20")
	src = strings.ReplaceAll(src, "_", "%5F")
	return src
}

func sanitizeQuotes(alt string) string {
	const replacementQuote = "'"
	for _, s := range []string{"\"", "“", "”", "&quot;"} {
		alt = strings.ReplaceAll(alt, s, replacementQuote)
	}
	return alt
}

func imageBlockReplacementFunction(groups []string) string {
	src := sanitizeLinks(groups[1])
	alt := ""
	var caption string

	if len(groups) > 2 {
		alt = groups[2]
	}
	if len(groups) > 3 {
		caption = groups[3]
		if alt == "" {
			alt = caption
		}
	} else {
		caption = alt
	}

	alt = sanitizeQuotes(alt)
	caption = sanitizeQuotes(caption)
	return fmt.Sprintf(`{{< figure src="%s" alt="%s" caption="%s" >}}`, src, alt, caption)
}

func captionReplacementFunction(groups []string) string {
	src := sanitizeLinks(groups[3])
	alt := ""

	if len(groups) > 4 {
		alt = sanitizeQuotes(groups[4])
	}

	return fmt.Sprintf(`{{< figure align="%s" width=%s src="%s" alt="%s" caption="%s" >}}`,
		groups[1], groups[2], src, alt, alt)
}
