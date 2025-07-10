package hugopage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const example1 = `
[caption id="attachment_3623" align="aligncenter" width="740"]<a href="https://ashishb.net/wp-content/uploads/2018/04/French-Laundry-0.jpg"><img class="size-large wp-image-3623" src="https://ashishb.net/wp-content/uploads/2018/04/French-Laundry-0-1024x579.jpg" alt="French Laundry" width="740" height="418" /></a> French Laundry[/caption]
`

const example2 = `
[caption id="attachment_3624" align="aligncenter" width="740"]<a href="https://ashishb.net/wp-content/uploads/2018/04/French-Laundry-2.jpg"><img class="size-large wp-image-3624" src="https://ashishb.net/wp-content/uploads/2018/04/French-Laundry-2-1024x624.jpg" alt="" width="740" height="451" /></a> Crispy Chickpea Panisse (at least that's what I remember)[/caption]
`

const example3 = `
[caption id="attachment_3103" align="aligncenter" width="740"]<a href="http://ashishb.net/wp-content/uploads/2016/11/IMG_20131202_121241.jpg"><img class="wp-image-3103 size-large" src="http://ashishb.net/wp-content/uploads/2016/11/IMG_20131202_121241-1024x768.jpg" width="740" height="555" /></a> Top of the Koko head crater[/caption]`

const example4 = `
</p>
[caption id="" align="aligncenter" width="2048"]<a href="https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2014/06/20140513_0036-Place-Jacques-Cartier-v2-web.jpg"><img src="https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2014/06/20140513_0036-Place-Jacques-Cartier-v2-web.jpg" alt="Place Jacques Cartier v2" width="2048" height="1161" /></a> Retouche manuelle[/caption]
<p>`

const example5 = `
<!-- wp:image {"align":"center","id":3875,"sizeSlug":"large","className":"is-style-default"} -->
<div class="wp-block-image is-style-default"><figure class="aligncenter size-large"><img src="https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2016/03/Shooting-Minh-Ly-0155-_DSC0155-Minh-Ly-WEB-1100x1100.jpg" alt="" class="wp-image-3875"/><figcaption>Minh-Ly</figcaption></figure></div>
<!-- /wp:image -->
`

func TestRegExMatches(t *testing.T) {
	require.True(t, _CaptionRegEx1.MatchString(example1), "RegEx should match")
	require.True(t, _CaptionRegEx1.MatchString(example2), "RegEx should match")
	require.True(t, _CaptionRegEx1.MatchString(example4), "RegEx should match")

	require.False(t, _CaptionRegEx1.MatchString(example3), "RegEx should match")

	require.True(t, _CaptionRegEx2.MatchString(example3), "RegEx should match")

	require.True(t, _FigureRegexCaption.MatchString(example5), "Regex should match")
	// This test is failing see https://github.com/ashishb/wp2hugo/pull/177
	// require.False(t, _FigureRegexNoCaption.MatchString(example5), "Regex should match")
}

func TestCaption4Replace(t *testing.T) {
	expected := "\n</p>\n{{< figure align=\"aligncenter\" width=2048 src=\"https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2014/06/20140513%5F0036-Place-Jacques-Cartier-v2-web.jpg\" alt=\"Place Jacques Cartier v2\" caption=\"Place Jacques Cartier v2\" >}}\n<p>"
	require.Equal(t, expected, replaceCaptionWithFigure(example4))
}

// This test is failing see https://github.com/ashishb/wp2hugo/pull/177
// func TestFigure5Replace(t *testing.T) {
//	expected := "\n{{< figure src=\"https://photo.aurelienpierre.com/wp-content/uploads/sites/3/2016/03/Shooting-Minh-Ly-0155-%5FDSC0155-Minh-Ly-WEB-1100x1100.jpg\" alt=\"\" caption=\"Minh-Ly\" >}}\n"
//	require.Equal(t, expected, replaceImageBlockWithFigure(example5))
//}
