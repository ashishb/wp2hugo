package hugopage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceGutenbergGallery(t *testing.T) {
	const htmlData = `<!-- wp:gallery {"ids":[14951,14949],"imageCrop":false,"linkTo":"file","sizeSlug":"full","align":"wide"} -->
<figure class="wp-block-gallery alignwide columns-2">
<ul class="blocks-gallery-grid">
<li class="blocks-gallery-item">
<figure><a href="https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2020/02/haute-diffusion-1.jpg">
<img src="https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2020/02/haute-diffusion-1.jpg" alt="" data-id="14951" data-full-url="https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2020/02/haute-diffusion-1.jpg" data-link="https://photo.aurelienpierre.com/la-photo-de-studio-pour-les-pauvres/haute-diffusion-1/" class="wp-image-14951"/></a>
<figcaption class="blocks-gallery-item__caption">Lumière fortement diffusée</figcaption>
</figure>
</li>
<li class="blocks-gallery-item">
<figure><a href="https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2020/02/faible-diffusion.jpg">
<img src="https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2020/02/faible-diffusion.jpg" alt="" data-id="14949" data-link="https://photo.aurelienpierre.com/la-photo-de-studio-pour-les-pauvres/faible-diffusion/" class="wp-image-14949"/></a>
<figcaption class="blocks-gallery-item__caption">Lumière faiblement diffusée<br /></figcaption>
</figure>
</li>
</ul>
</figure>
<!-- /wp:gallery -->`
	const expected = `<br>{{< gallery cols="2" >}}<br>{{< figure src="https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2020/02/haute-diffusion-1.jpg" alt="Lumière fortement diffusée" caption="Lumière fortement diffusée" >}}<br>{{< figure src="https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2020/02/faible-diffusion.jpg" alt="Lumière faiblement diffusée<br/>" caption="Lumière faiblement diffusée<br/>" >}}<br>{{< /gallery >}}<br>`
	assert.Equal(t, expected, replaceGutembergGalleryWithFigure(htmlData))
}
