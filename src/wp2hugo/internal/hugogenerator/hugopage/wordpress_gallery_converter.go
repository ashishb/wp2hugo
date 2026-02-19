package hugopage

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

// Example:  [gallery size="medium" link="file" columns="4" ids="1710,1713,1712,1711"]
// Full examples : https://codex.wordpress.org/Gallery_Shortcode
// the important field to extract is "ids", but others might come in handy
// `size` is legacy from pre-responsive design and should be discarded now
// `link` is probably something to enforce in Hugo figure shortcode,
// It is mostly "file" to handle, since "attachment_page" makes no sense for Hugo.
var (
	_GalleryRegEx = regexp.MustCompile(`\[gallery ([^\[\]]*)\]`)
	_idRegEx      = regexp.MustCompile(`ids="([^"]+)"`)
	_colsRegEx    = regexp.MustCompile(`columns="([^"]+)"`)
)

// Example:
// <!-- wp:gallery {"ids":[14951,14949],"imageCrop":false,"linkTo":"file","sizeSlug":"full","align":"wide"} -->
// <figure class="wp-block-gallery alignwide columns-2"><ul class="blocks-gallery-grid"><li class="blocks-gallery-item"><figure><a href="https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2020/02/haute-diffusion-1.jpg"><img src="https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2020/02/haute-diffusion-1.jpg" alt="" data-id="14951" data-full-url="https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2020/02/haute-diffusion-1.jpg" data-link="https://photo.aurelienpierre.com/la-photo-de-studio-pour-les-pauvres/haute-diffusion-1/" class="wp-image-14951"/></a><figcaption class="blocks-gallery-item__caption">Lumière fortement diffusée</figcaption></figure></li><li class="blocks-gallery-item"><figure><a href="https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2020/02/faible-diffusion.jpg"><img src="https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2020/02/faible-diffusion.jpg" alt="" data-id="14949" data-link="https://photo.aurelienpierre.com/la-photo-de-studio-pour-les-pauvres/faible-diffusion/" class="wp-image-14949"/></a><figcaption class="blocks-gallery-item__caption">Lumière faiblement diffusée<br /></figcaption></figure></li></ul></figure>
// <!-- /wp:gallery -->
var _GutenbergGalleryRegEx = regexp.MustCompile(`(?ms)<!-- wp:gallery.*?-->(.*?)<!-- /wp:gallery -->`)

var _innerFigureNoCaption = regexp.MustCompile(`(?ms)<figure.*?` +
	`<img.*?src="([^"]+)".*?alt="([^"]*)".*?>.*?` +
	`.*?</figure>`)

var _innerFigureCaption = regexp.MustCompile(`(?ms)<figure.*?` +
	`<img.*?src="([^"]+)".*?alt="([^"]*)".*?>.*?` +
	`<figcaption.*?>(.*?)</figcaption>` +
	`.*?</figure>`)

var errGalleryWithNoIDs = errors.New("no image IDs found in gallery shortcode")

// TODO: should we handle `order="ASC|DESC"` when `orderby="ID"` ?
// Seems to me that people mostly order pictures in galleries arbitrarily.
// Converts the WordPress's caption shortcode to Hugo shortcode "figure"
func replaceGalleryWithFigure(provider ImageURLProvider, attachmentIDs []string, htmlData string) string {
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

func replaceGutembergGalleryWithFigure(htmlData string) string {
	log.Debug().
		Msg("Replacing Gutenberg gallery with figures")

	return replaceAllStringSubmatchFunc(_GutenbergGalleryRegEx, htmlData, gutenbergGalleryReplacementFunction)
}

// Recursively find <figure> nodes
func findInnerFigures(node *html.Node, results *[]*html.Node) {
	if node.Type == html.ElementNode && node.Data == "figure" {
		*results = append(*results, node)
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		findInnerFigures(c, results)
	}
}

func renderNode(n *html.Node) string {
	var b strings.Builder
	if err := html.Render(&b, n); err != nil {
		log.Warn().Err(err).Msg("Failed to render HTML node")
		return ""
	} else {
		return b.String()
	}
}

func replaceGalleryFigure(htmlData string) string {
	log.Debug().
		Msg("Replacing Gutenberg image with figure")

	htmlData = replaceAllStringSubmatchFunc(_innerFigureCaption, htmlData, imageBlockReplacementFunction)
	htmlData = replaceAllStringSubmatchFunc(_innerFigureNoCaption, htmlData, imageBlockReplacementFunction)
	return htmlData
}

func gutenbergGalleryReplacementFunction(groups []string) string {
	// Because <figure> elements can be recursively nested,
	// we can't use RegEx, we need an HTML parser.
	doc, err := html.Parse(strings.NewReader(groups[1]))
	if err != nil {
		return groups[0]
	}
	// Produce a flat list of (possibly nested) figures
	var figures []*html.Node
	findInnerFigures(doc, &figures)
	cols := "1"
	inners := make([]string, 0)

	for _, f := range figures {
		isInner := false

		// Gutenberg gallery blocks have a top-level <figure>
		// holding the CSS classes for styling, and <figure> children containing the captions,
		// which should not contain classes. That's how we try to guess which is which.
		classAttr := ""
		for _, attr := range f.Attr {
			if attr.Key == "class" {
				classAttr = attr.Val
				break
			}
		}

		if classAttr != "" {
			re := regexp.MustCompile(`columns-(\d+)`)
			matches := re.FindStringSubmatch(classAttr)
			if len(matches) == 2 {
				cols = matches[1]
			} else {
				isInner = true
			}
		} else {
			isInner = true
		}

		// If we have an inner figure, parse it with our standard methods
		if isInner {
			inners = append(inners, replaceGalleryFigure(renderNode(f)))
		}
	}

	var output strings.Builder
	output.WriteString("<br>") // This will get converted to newline later on
	fmt.Fprintf(&output, `{{< gallery cols="%s" >}}`, cols)
	output.WriteString("<br>") // This will get converted to newline later on

	for _, f := range inners {
		output.WriteString(f)
		output.WriteString("<br>") // This will get converted to newline later on
	}
	output.WriteString(`{{< /gallery >}}`)
	output.WriteString("<br>") // This will get converted to newline later on

	return output.String()
}

func galleryReplacementFunction(provider ImageURLProvider, attachmentIDs []string, galleryInfo string) (string, error) {
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
			ids = []string{"", strings.Join(attachmentIDs, ",")}
			log.Info().
				Str("galleryInfo", galleryInfo).
				Strs("attachmentIDs", attachmentIDs).
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
	fmt.Fprintf(&output, `{{< gallery cols="%s" >}}`, colNb)

	// For each image ID in WP gallery shortcode, get the URL
	for _, s := range idsArray {
		tmp, err := provider.GetImageInfo(s)
		if tmp != nil {
			src := sanitizeLinks(tmp.ImageURL)
			title := sanitizeQuotes(tmp.Title)

			output.WriteString("<br>") // This will get converted to newline later on
			fmt.Fprintf(&output, `{{< figure src="%s" title="%s" alt="%s" >}}`, src, title, title)
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
