package imagealtsuggest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_regex1(t *testing.T) {
	t.Parallel()
	const input = `something {{< figure align=aligncenter width=740 src="IMG_0384-1-1024x768.jpg" alt="" >}}`
	const expectedSrc = "IMG_0384-1-1024x768.jpg"
	const expectedAlt = ""
	matches := _figureShortCodeRegEx.FindAllStringSubmatch(input, -1)
	require.Len(t, matches, 1)
	srcMatches := _figureShortCodeSrcRegEx.FindAllStringSubmatch(matches[0][0], -1)
	require.Len(t, srcMatches, 1)
	require.Len(t, srcMatches[0], 2)
	require.Equal(t, expectedSrc, srcMatches[0][1])

	altMatches := _figureShortCodeAltRegEx.FindAllStringSubmatch(matches[0][0], -1)
	require.Len(t, altMatches, 1)
	require.Len(t, altMatches[0], 2)
	require.Equal(t, expectedAlt, altMatches[0][1])
}
