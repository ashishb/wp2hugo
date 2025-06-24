package hugopage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testCases = []string{
	`[nk_awb awb_type="image" awb_image="4256" awb_stretch="true" awb_image_size="full" awb_image_background_size="cover" awb_image_background_position="50% 50%" awb_parallax="scroll-opacity" awb_parallax_speed="0.5" awb_parallax_mobile="true"]`,
	`[nk_awb awb_type="image" awb_image="2813" awb_stretch="true" awb_image_size="full" awb_image_background_size="cover" awb_image_background_position="40% 40%" awb_parallax="scroll-opacity" awb_parallax_speed="0.5" awb_parallax_mobile="true"]`,
	`[nk_awb awb_type="image" awb_image="3992" awb_stretch="true" awb_image_size="full" awb_image_background_size="cover" awb_image_background_position="50% 50%" awb_parallax="scroll-opacity" awb_parallax_speed="0.5" awb_parallax_mobile="true"]`,
	`[nk_awb awb_type="image" awb_stretch="true" awb_image="3517" awb_image_size="awb_xl" awb_image_background_size="cover" awb_image_background_position="50% 50%" awb_parallax="scroll-opacity" awb_parallax_speed="0.5" awb_parallax_mobile="true"]`,
}

func TestParallelBlurRegEx(t *testing.T) {
	for _, testCase := range testCases {
		assert.True(t, _AWBRegEx.MatchString(testCase))
	}
}
