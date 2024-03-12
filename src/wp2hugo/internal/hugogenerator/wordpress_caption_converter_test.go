package hugogenerator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const example1 = `
[caption id="attachment_3623" align="aligncenter" width="740"]<a href="https://ashishb.net/wp-content/uploads/2018/04/French-Laundry-0.jpg"><img class="size-large wp-image-3623" src="https://ashishb.net/wp-content/uploads/2018/04/French-Laundry-0-1024x579.jpg" alt="French Laundry" width="740" height="418" /></a> French Laundry[/caption]
`

const example2 = `
[caption id="attachment_3624" align="aligncenter" width="740"]<a href="https://ashishb.net/wp-content/uploads/2018/04/French-Laundry-2.jpg"><img class="size-large wp-image-3624" src="https://ashishb.net/wp-content/uploads/2018/04/French-Laundry-2-1024x624.jpg" alt="" width="740" height="451" /></a> Crispy Chickpea Panisse (at least that's what I remember)[/caption]
`

const example3 = `
[caption id="attachment_3103" align="aligncenter" width="740"]<a href="http://ashishb.net/wp-content/uploads/2016/11/IMG_20131202_121241.jpg"><img class="wp-image-3103 size-large" src="http://ashishb.net/wp-content/uploads/2016/11/IMG_20131202_121241-1024x768.jpg" width="740" height="555" /></a> Top of the Koko head crater[/caption]
`

func TestRegExMatches(t *testing.T) {
	assert.True(t, _CaptionRegEx1.MatchString(example1), "RegEx should match")
	assert.True(t, _CaptionRegEx1.MatchString(example2), "RegEx should match")

	assert.False(t, _CaptionRegEx1.MatchString(example3), "RegEx should match")

	assert.True(t, _CaptionRegEx2.MatchString(example3), "RegEx should match")
}
