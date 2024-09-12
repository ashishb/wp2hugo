package hugogenerator

import (
	"errors"
	"fmt"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/rs/zerolog/log"
	"path"
)

// This will go to  layouts/shortcodes/googlemaps.html enabling the use of the shortcode
// "googlemaps" in the markdown files
const _googleMapsShortCode = `<iframe loading="lazy"
        src="https://www.google.com/maps/d/embed?mid={{ .Get "src" }}"
        width="{{ .Get "width" }}"
        height="{{ .Get "height" }}">
</iframe>
`

const _selectedPostsShortCode = `{{ $category := .Get "category" }}
{{ $catLink := .Get "catlink" | default true }}
{{ $count := .Get "count" | default 5 }}

{{ $p := site.AllPages }}
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

// converter for https://wordpress.org/plugins/advanced-backgrounds/
const _ParallaxBlurShortCode = `{{ $imgURL := .Get "src" }}
{{ $id := substr (md5 .Inner) 0 16 }}

<style>
#div-{{$id}} {
    position: relative; /* Allows layering elements */
    height: auto; /* Adjust height as needed */
}

#div-{{$id}}:after {
    content: "";
    position: absolute; /* Overlays content */
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-image: url("{{ $imgURL }}");
    background-attachment: fixed;
    background-size: cover;
    background-position: center;
    filter: blur(0px); /* Initial blur */
    transition: filter 0.5s ease; /* Smooth transition */
}

#div-{{$id}} p {
    text-align: center;
}


#div-{{$id}} a {
    color: white; /* Adjust text color for contrast */
    text-align: center;
    text-shadow: 0 3px 0 gray;
    position: relative; /* Allows text to stay above blur */
    z-index: 1; /* Ensures text is on top of blurred background */
    backdrop-filter: blur(5px); /* Amount of blur on text */
    padding: 0.5rem;
}

#div-{{$id}} a:link,a:visited,a:hover,a:active {
   text-decoration: none !important;
   text-decoration-style: unset !important;
   text-decoration-thickness: 0 !important;
}

/* Increase blur on scroll */
#div-{{$id}}:after {
    opacity: 0.8; /* Semi-transparent background */
}

#div-{{$id}}:hover:after,
#div-{{$id}}:active:after,
#div-{{$id}}:focus:after,
#div-{{$id}}:target:after {
    filter: blur(10px); /* Amount of blur on interaction */
}
</style>

<div class="container" id="div-{{ $id }}">
    {{ .Inner | markdownify }}
</div>
`

const _audioShortCode = `
<audio controls preload="metadata">
  <source src="{{ .Get "src" }}" type="audio/{{ replace (path.Ext (.Get "src")) "." ""}}">
  Your browser does not support the audio element.
</audio>
`

func WriteCustomShortCodes(siteDir string) error {
	return errors.Join(writeGoogleMapsShortCode(siteDir),
		writeSelectedPostsShortCode(siteDir),
		writeParallaxBlurShortCode(siteDir),
		writeAudioShortCode(siteDir))
}

func writeGoogleMapsShortCode(siteDir string) error {
	return writeShortCode(siteDir, "googlemaps", _googleMapsShortCode)
}

func writeSelectedPostsShortCode(siteDir string) error {
	return writeShortCode(siteDir, "catlist", _selectedPostsShortCode)
}

func writeParallaxBlurShortCode(siteDir string) error {
	return writeShortCode(siteDir, "parallaxblur", _ParallaxBlurShortCode)
}

func writeAudioShortCode(siteDir string) error {
	return writeShortCode(siteDir, "audio", _audioShortCode)
}

func writeShortCode(siteDir string, shortCodeName string, fileContent string) error {
	log.Debug().
		Str("shortcode", shortCodeName).
		Msg("Writing shortcode")
	shortCodeDir := path.Join(siteDir, "layouts", "shortcodes")
	if err := utils.CreateDirIfNotExist(path.Join(siteDir, "layouts", "shortcodes")); err != nil {
		return err
	}
	googleMapsFile := path.Join(shortCodeDir, fmt.Sprintf("%s.html", shortCodeName))
	return writeFile(googleMapsFile, []byte(fileContent))
}
