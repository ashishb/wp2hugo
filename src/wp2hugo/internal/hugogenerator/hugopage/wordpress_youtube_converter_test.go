package hugopage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplaceYoutubeURL1(t *testing.T) {
	t.Parallel()
	const htmlData = "This is a test with a youtube link:\nhttps://www.youtube.com/watch?v=gL0-m1Qlohg"
	const expected = "This is a test with a youtube link:\n{{< youtube gL0-m1Qlohg >}}"
	require.Equal(t, expected, replacePlaintextYoutubeURL(htmlData))
}

func TestReplaceYoutubeURL2(t *testing.T) {
	t.Parallel()
	const htmlData = "This is a test with a youtube link: https://www.youtube.com/watch?v=8K7PdBH3W_I"
	const expected = "This is a test with a youtube link: {{< youtube 8K7PdBH3W_I >}}"
	require.Equal(t, expected, replacePlaintextYoutubeURL(htmlData))
}

func TestReplaceYoutubeURL3(t *testing.T) {
	t.Parallel()
	const htmlData = "This is a test with a youtube link:\thttps://www.youtube.com/watch?v=gJ7AAJXHeeg whatever"
	const expected = "This is a test with a youtube link:\t{{< youtube gJ7AAJXHeeg >}} whatever"
	require.Equal(t, expected, replacePlaintextYoutubeURL(htmlData))
}

// This test is failing see https://github.com/ashishb/wp2hugo/pull/177
// func TestReplaceYoutubeURL4(t *testing.T) {
//	const htmlData = "[embed]https://www.youtube.com/watch?v=gJ7AAJXHeeg[/embed]"
//	const expected = "{{< youtube gJ7AAJXHeeg >}}"
//	require.Equal(t, expected, replacePlaintextYoutubeURL(htmlData))
//}

func TestReplaceNonPlaintextYouTubeURL(t *testing.T) {
	t.Parallel()
	const htmlData = `This is a test with a youtube link <a href="https://www.youtube.com/watch?v=gJ7AAJXHeeg" and
		embed <iframe width="560" height="315" src="https://www.youtube.com/embed/Wz6ml5SpkKM?si=rrx_5_80TE3Mz7Co"
		title="YouTube video player" frameborder="0" allowfullscreen></iframe>`
	// Assert that the function does not replace the youtube URL in the iframe or the link
	require.Equal(t, htmlData, replacePlaintextYoutubeURL(htmlData))
}

func TestReplaceYoutubeURL5(t *testing.T) {
	t.Parallel()
	const htmlData = `<!-- wp:embed {"url":"https://www.youtube.com/watch?v=7l6FjphZXsk","type":"video","providerNameSlug":"youtube","responsive":true,"align":"full","className":"wp-embed-aspect-16-9 wp-has-aspect-ratio"} -->
		<figure class="wp-block-embed alignfull is-type-video is-provider-youtube wp-block-embed-youtube wp-embed-aspect-16-9 wp-has-aspect-ratio"><div class="wp-block-embed__wrapper">
		https://www.youtube.com/watch?v=7l6FjphZXsk
	</div></figure>
	<!-- /wp:embed -->`
	const expected = "{{< youtube 7l6FjphZXsk >}}"
	require.Equal(t, expected, replacePlaintextYoutubeURL(htmlData))
}
