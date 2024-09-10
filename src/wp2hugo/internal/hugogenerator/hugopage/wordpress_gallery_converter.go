package hugopage

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// Example:  [gallery size="medium" link="file" columns="4" ids="1710,1713,1712,1711"]
// Full examples : https://codex.wordpress.org/Gallery_Shortcode
// the important field to extract is "ids", but others might come in handy
// `size` is legacy from pre-responsive design and should be discarded now
// `link` is probably something to enforce in Hugo figure shortcode,
// for us it's mostly "file" to handle, since "attachment_page" makes no sense for Hugo.
var _GalleryRegEx = regexp.MustCompile(`\[gallery ([^\[\]]*)\]`)

var _idRegEx = regexp.MustCompile(`ids="([^"]+)"`)
var _colsRegEx = regexp.MustCompile(`columns="([^"]+)"`)

// TODO: should we handle `order="ASC|DESC"` when `orderby="ID"` ?
// Seems to me that people mostly order pictures in galleries arbitrarily.

// Converts the WordPress's caption shortcode to Hugo shortcode "figure"
// https://adityatelange.github.io/hugo-PaperMod/posts/papermod/papermod-faq/#centering-image-in-markdown
func replaceGalleryWithFigure(provider ImageURLProvider, htmlData string) string {
	log.Debug().
		Msg("Replacing gallery with figures")

	htmlData = replaceAllStringSubmatchFunc(_GalleryRegEx, htmlData,
		func(groups []string) string {
			return galleryReplacementFunction(provider, groups)
		})

	return htmlData
}

func galleryReplacementFunction(provider ImageURLProvider, groups []string) string {
	var output strings.Builder

	// Find columns layout
	cols := _colsRegEx.FindStringSubmatch(groups[1])
	col_nb := "1"
	if cols != nil {
		col_nb = cols[1]
	}

	// Find image IDs
	ids := _idRegEx.FindStringSubmatch(groups[1])
	ids_array := strings.Split(ids[1], ",")

	// TODO: maybe handle `order="ASC|DESC"` in conjunction with `orderby="..."`, so reorder ids_array here.

	// We will use `figure` shortcodes nested into a `gallery` shortcode for the main layout
	output.WriteString("<br>") // This will get converted to newline later on

	// Opening tag :
	output.WriteString(fmt.Sprintf(`{{%% gallery cols="%s" %%}}`,  col_nb))

	// For each image ID in WP gallery shortcode, get the URL
	for _, s := range ids_array {
		tmp, err := provider.GetImageInfo(s)
		if tmp != nil {
			src := tmp.ImageURL
			// These characters create problems in Hugo's markdown
			src = strings.ReplaceAll(src, " ", "%20")
			src = strings.ReplaceAll(src, "_", "%5F")

			title := tmp.Title

			output.WriteString("<br>") // This will get converted to newline later on
			output.WriteString(fmt.Sprintf(`{{< figure src="%s" title="%s" alt="%s" >}}`, src, title, title))
			output.WriteString("<br>") // This will get converted to newline later on
		} else {
			log.Warn().
				Err(err).
				Str("imageID", s).
				Msg("Image URL not found")
		}
	}

	// Closing tag for main gallery shortcode
	output.WriteString(`{{% /gallery %}}`)
	output.WriteString("<br>") // This will get converted to newline later on
	return output.String()
}
