package hugogenerator

import (
	"github.com/rs/zerolog/log"
	"path"
)

// This will go to  layouts/shortcodes/googlemaps.html enabling the use of the shortcode
// "googlemaps" in the markdown files
const _googleMapsShortCode = `
<iframe loading="lazy"
        src="https://www.google.com/maps/d/embed?mid={{ .Get "src" }}"
        width="{{ .Get "width" }}"
        height="{{ .Get "height" }}">
</iframe>
`

func writeCustomShortCodes(siteDir string) error {
	return writeGoogleMapsShortCode(siteDir)
}

func writeGoogleMapsShortCode(siteDir string) error {
	log.Debug().Msg("Writing googlemaps shortcode")
	shortCodeDir := path.Join(siteDir, "layouts", "shortcodes")
	if err := createDirIfNotExist(path.Join(siteDir, "layouts")); err != nil {
		return err
	}
	if err := createDirIfNotExist(path.Join(siteDir, "layouts", "shortcodes")); err != nil {
		return err
	}
	googleMapsFile := path.Join(shortCodeDir, "googlemaps.html")
	return writeFile(googleMapsFile, []byte(_googleMapsShortCode))
}
