package hugopage

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"regexp"
	"strings"
)

type ImageURLProvider interface {
	// E.g. converts "4256" to "https://ashishb.net/wp-content/uploads/2018/12/bora_bora_5_resized.jpg"
	GetImageURL(imageID string) (*string, error)
}

// Example: [nk_awb awb_type="image" awb_image="4256" awb_stretch="true" awb_image_size="full" awb_image_background_size="cover" awb_image_background_position="50% 50%" awb_parallax="scroll-opacity" awb_parallax_speed="0.5" awb_parallax_mobile="true"]
var _AWBRegEx = regexp.MustCompile(`\[nk_awb awb_type="image".*?awb_image="([^"]+)".*?]`)

// Partial converter for https://wordpress.org/plugins/advanced-backgrounds/
func replaceAWBWithParallaxBlur(provider ImageURLProvider, htmlData string) string {
	log.Debug().
		Msg("Replacing AWB (Advanced WordPress Backgrounds) with parallaxblur")

	htmlData = replaceAllStringSubmatchFunc(_AWBRegEx, htmlData,
		func(groups []string) string {
			return awbReplacementFunction(provider, groups)
		})
	htmlData = strings.ReplaceAll(htmlData, "[/nk_awb]", "{{< /parallaxblur >}}")
	return htmlData
}

func awbReplacementFunction(provider ImageURLProvider, groups []string) string {
	srcImageID := groups[1]
	tmp, err := provider.GetImageURL(srcImageID)
	if tmp == nil {
		log.Fatal().
			Err(err).
			Str("imageID", srcImageID).
			Msg("Image URL not found")
	}
	src := *tmp
	// These character creates problem in Hugo's markdown
	src = strings.ReplaceAll(src, " ", "%20")
	src = strings.ReplaceAll(src, "_", "%5F")
	return fmt.Sprintf(`{{< parallaxblur src="%s" >}}`, src)
}
