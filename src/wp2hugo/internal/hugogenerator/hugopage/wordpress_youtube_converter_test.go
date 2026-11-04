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

func TestReplaceYoutubeURL4(t *testing.T) {
	t.Parallel()
	const htmlData = "[embed]https://www.youtube.com/watch?v=gJ7AAJXHeeg[/embed]"
	const expected = "{{< youtube gJ7AAJXHeeg >}}"
	require.Equal(t, expected, replacePlaintextYoutubeURL(htmlData))
}

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

func TestReplaceYoutubeURL6(t *testing.T) {
	t.Parallel()
	// Ref: https://github.com/ashishb/wp2hugo/issues/268
	const htmlData = "[youtube http://www.youtube.com/watch?v=1cNDSPutas8]"
	const expected = "{{< youtube 1cNDSPutas8 >}}"
	require.Equal(t, expected, replacePlaintextYoutubeURL(htmlData))
}

func TestReplaceYoutubeShortURL(t *testing.T) {
	t.Parallel()
	const htmlData = "Intro\nhttps://youtu.be/gJ7AAJXHeeg\nOutro"
	const expected = "Intro\n{{< youtube gJ7AAJXHeeg >}}\nOutro"
	require.Equal(t, expected, replacePlaintextYoutubeURL(htmlData))
}

func TestReplaceYoutubeURLWithQueryParameters(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"timestamp", "https://www.youtube.com/watch?v=gJ7AAJXHeeg&t=30s", "{{< youtube gJ7AAJXHeeg >}}"},
		{"short URL with share id", "https://youtu.be/gJ7AAJXHeeg?si=AbC_123", "{{< youtube gJ7AAJXHeeg >}}"},
		{"short URL with timestamp", "https://youtu.be/gJ7AAJXHeeg?t=42 whatever", "{{< youtube gJ7AAJXHeeg >}} whatever"},
		{"trailing period", "See https://youtu.be/gJ7AAJXHeeg.", "See {{< youtube gJ7AAJXHeeg >}}."},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, replacePlaintextYoutubeURL(tc.input))
		})
	}
}

func TestReplaceYoutubeURLWithEscapedUnderscore(t *testing.T) {
	t.Parallel()
	// The HTML to Markdown converter escapes "_" as "\_"
	const htmlData = `https://www.youtube.com/watch?v=8K7PdBH3W\_I`
	const expected = "{{< youtube 8K7PdBH3W_I >}}"
	require.Equal(t, expected, replacePlaintextYoutubeURL(htmlData))
}

func TestReplaceGutenbergYoutubeEmbedWithTimestamp(t *testing.T) {
	t.Parallel()
	const htmlData = `<!-- wp:embed {"url":"https://www.youtube.com/watch?v=7l6FjphZXsk\u0026t=30s","type":"video","providerNameSlug":"youtube"} -->
	<figure class="wp-block-embed is-type-video is-provider-youtube wp-block-embed-youtube"><div class="wp-block-embed__wrapper">
	https://www.youtube.com/watch?v=7l6FjphZXsk&t=30s
	</div></figure>
	<!-- /wp:embed -->`
	const expected = "{{< youtube 7l6FjphZXsk >}}"
	require.Equal(t, expected, replacePlaintextYoutubeURL(htmlData))
}

func TestYoutubeURLsInPageContent(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name         string
		htmlInput    string
		expectedBody string
	}{
		{"short URL", "<p>Intro</p>\n<p>https://youtu.be/gJ7AAJXHeeg</p>", "Intro\n\n{{< youtube gJ7AAJXHeeg >}}"},
		{"timestamp", "<p>Intro</p>\n<p>https://www.youtube.com/watch?v=gJ7AAJXHeeg&t=30s</p>", "Intro\n\n{{< youtube gJ7AAJXHeeg >}}"},
		{"underscore in ID", "<p>https://www.youtube.com/watch?v=8K7PdBH3W_I</p>", "{{< youtube 8K7PdBH3W_I >}}"},
		{
			"Gutenberg embed with short URL",
			`<!-- wp:embed {"url":"https://youtu.be/gJ7AAJXHeeg","type":"video","providerNameSlug":"youtube"} -->
<figure class="wp-block-embed is-type-video is-provider-youtube wp-block-embed-youtube"><div class="wp-block-embed__wrapper">
https://youtu.be/gJ7AAJXHeeg
</div></figure>
<!-- /wp:embed -->`,
			"{{< youtube gJ7AAJXHeeg >}}",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testMarkdownExtractor(t, tc.htmlInput, tc.expectedBody)
		})
	}
}
