package hugogenerator

import (
	"errors"
	"fmt"
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

const _selectedPostsShortCode = `
{{ $category := .Get "category" }}
{{ $catLink := .Get "catlink" | default true }}
{{ $count := .Get "count" | default 5 }}

{{ $p := where site.RegularPages "Type" "posts" }}
{{ $p = where $p "Params.category" "intersect" (slice $category) }}

{{ $categoryTitle := title $category }}
{{ $categoryTitle = strings.Replace $categoryTitle "-" " " }}

<h3>
  {{ if $catLink }}
    <a href="/category/{{ urlquery $category }}"> {{$categoryTitle }} </a>
  {{ else }}
    {{ $categoryTitle }}
  {{ end }}

</h3>
<ul>
  {{ range first $count $p}}
      <li><a href="{{ .RelPermalink }}">{{ .Title }}</a>
      </li>
  {{ end }}
</ul>
`

func WriteCustomShortCodes(siteDir string) error {
	err1 := writeGoogleMapsShortCode(siteDir)
	err2 := writeSelectedPostsShortCode(siteDir)
	return errors.Join(err1, err2)
}

func writeGoogleMapsShortCode(siteDir string) error {
	return writeShortCode(siteDir, "googlemaps", _googleMapsShortCode)
}

func writeSelectedPostsShortCode(siteDir string) error {
	return writeShortCode(siteDir, "catlist", _selectedPostsShortCode)
}

func writeShortCode(siteDir string, shortCodeName string, fileContent string) error {
	log.Debug().
		Str("shortcode", shortCodeName).
		Msg("Writing shortcode")
	shortCodeDir := path.Join(siteDir, "layouts", "shortcodes")
	if err := createDirIfNotExist(path.Join(siteDir, "layouts", "shortcodes")); err != nil {
		return err
	}
	googleMapsFile := path.Join(shortCodeDir, fmt.Sprintf("%s.html", shortCodeName))
	return writeFile(googleMapsFile, []byte(fileContent))
}
