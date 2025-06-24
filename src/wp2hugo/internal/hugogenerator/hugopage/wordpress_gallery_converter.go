package hugopage

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// Example:  [gallery size="medium" link="file" columns="4" ids="1710,1713,1712,1711"]
// Full examples : https://codex.wordpress.org/Gallery_Shortcode
// the important field to extract is "ids", but others might come in handy
// `size` is legacy from pre-responsive design and should be discarded now
// `link` is probably something to enforce in Hugo figure shortcode,
// It is mostly "file" to handle, since "attachment_page" makes no sense for Hugo.
var _GalleryRegEx = regexp.MustCompile(`\[gallery ([^\[\]]*)\]`)

var _idRegEx = regexp.MustCompile(`ids="([^"]+)"`)
var _colsRegEx = regexp.MustCompile(`columns="([^"]+)"`)

var errGalleryWithNoIDs = errors.New("no image IDs found in gallery shortcode")

// TODO: should we handle `order="ASC|DESC"` when `orderby="ID"` ?
// Seems to me that people mostly order pictures in galleries arbitrarily.

// Converts the WordPress's caption shortcode to Hugo shortcode "figure"
// https://adityatelange.github.io/hugo-PaperMod/posts/papermod/papermod-faq/#centering-image-in-markdown
func replaceGalleryWithFigure(provider ImageURLProvider, attachmentIDs []int, htmlData string) string {
	log.Debug().
		Msg("Replacing gallery with figures")

	htmlData = replaceAllStringSubmatchFunc(_GalleryRegEx, htmlData,
		func(groups []string) string {
			info, err := galleryReplacementFunction(provider, attachmentIDs, groups[1])
			if err != nil {
				return fmt.Sprintf("[gallery %s]", info) // Return the original shortcode
			}
			return info
		})

	return htmlData
}

func galleryReplacementFunction(provider ImageURLProvider, attachmentIDs []int, galleryInfo string) (string, error) {
	var output strings.Builder

	// Find columns layout
	cols := _colsRegEx.FindStringSubmatch(galleryInfo)
	colNb := "1"
	if cols != nil {
		colNb = cols[1]
	}

	// Find image IDs
	ids := _idRegEx.FindStringSubmatch(galleryInfo)
	if len(ids) == 0 {
		if len(attachmentIDs) > 0 {
			idsStr := make([]string, len(attachmentIDs))
			for i, id := range attachmentIDs {
				idsStr[i] = fmt.Sprintf("%d", id)
			}
			ids = []string{"", strings.Join(idsStr, ",")}
			log.Info().
				Str("galleryInfo", galleryInfo).
				Ints("attachmentIDs", attachmentIDs).
				Msg("No image IDs found in gallery shortcode, fallback to page attachments")
		} else {
			log.Warn().
				Msg("No image IDs found in gallery shortcode and no page attachments")
			return "", errGalleryWithNoIDs
		}
	}

	idsArray := strings.Split(ids[1], ",")

	// TODO: maybe handle `order="ASC|DESC"` in conjunction with `orderby="..."`, so reorder ids_array here.

	// We will use `figure` shortcodes nested into a `gallery` shortcode for the main layout
	output.WriteString("<br>") // This will get converted to newline later on

	// Opening tag :
	output.WriteString(fmt.Sprintf(`{{< gallery cols="%s" >}}`, colNb))

	// For each image ID in WP gallery shortcode, get the URL
	for _, s := range idsArray {
		imgID, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			log.Warn().
				Err(err).
				Str("imageID", s).
				Msg("Invalid image ID in gallery")
			continue
		}
		tmp, err := provider.GetImageInfo(imgID)
		if tmp != nil {
			src := tmp.ImageURL
			// These characters create problems in Hugo's markdown
			src = strings.ReplaceAll(src, " ", "%20")
			src = strings.ReplaceAll(src, "_", "%5F")

			// Escape weird characters in title
			title := strings.ReplaceAll(tmp.Title, `"`, `\"`)
			title = strings.ReplaceAll(title, "\n", " ")

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
	output.WriteString(`{{< /gallery >}}`)
	output.WriteString("<br>") // This will get converted to newline later on
	return output.String(), nil
}
