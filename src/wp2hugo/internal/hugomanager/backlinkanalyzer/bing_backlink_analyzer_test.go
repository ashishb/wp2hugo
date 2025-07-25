package backlinkanalyzer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_BacklinkURL_Normalize(t *testing.T) {
	t.Parallel()
	b1, err := BacklinkURL("https://play.google.com/store/apps/details?hl=or&id=net.ashishb.androidmusicplayer").Normalize()
	require.NoError(t, err)
	require.Equal(t, "https://play.google.com/store/apps/details?id=net.ashishb.androidmusicplayer", string(*b1))

	b2, err := BacklinkURL("https://f-droid.org/bo/2025/06/26/twif.html").Normalize()
	require.NoError(t, err)
	require.Equal(t, "https://f-droid.org/2025/06/26/twif.html", string(*b2))

	b3, err := BacklinkURL("https://f-droid.org/zh_hant/2025/06/26/twif.html").Normalize()
	require.NoError(t, err)
	require.Equal(t, "https://f-droid.org/2025/06/26/twif.html", string(*b3))
}
