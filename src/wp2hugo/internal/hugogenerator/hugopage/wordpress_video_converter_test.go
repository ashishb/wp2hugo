package hugopage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplaceVideo1(t *testing.T) {
	t.Parallel()
	const htmlData = `[video src="/wp-content/uploads/2026/05/video.mp4"][/video]`
	const expected = `{{< video src="/wp-content/uploads/2026/05/video.mp4" >}}`
	require.Equal(t, expected, replaceVideoShortCode(htmlData))
}

func TestReplaceVideo2(t *testing.T) {
	t.Parallel()
	const htmlData = `[video mp4="/wp-content/uploads/2026/05/video.mp4"]`
	const expected = `{{< video src="/wp-content/uploads/2026/05/video.mp4" >}}`
	require.Equal(t, expected, replaceVideoShortCode(htmlData))
}

func TestReplaceVideo3(t *testing.T) {
	t.Parallel()
	const htmlData = `[video mp4="/wp-content/uploads/2026/05/video.mp4" webm="/wp-content/uploads/2026/05/video.webm"]`
	const expected = `{{< video src="/wp-content/uploads/2026/05/video.mp4" >}}`
	require.Equal(t, expected, replaceVideoShortCode(htmlData))
}

func TestReplaceVideo4(t *testing.T) {
	t.Parallel()
	const htmlData = `<figure class="wp-block-video"><video controls src="/wp-content/uploads/2026/05/video.mp4"></video></figure>`
	const expected = `{{< video src="/wp-content/uploads/2026/05/video.mp4" >}}`
	require.Equal(t, expected, replaceVideoShortCode(htmlData))
}

func TestReplaceVideo5(t *testing.T) {
	t.Parallel()
	const htmlData = `<figure class="wp-block-video aligncenter"><video controls src="/wp-content/uploads/2026/05/my_video.mp4"></video>
	</figure>`
	const expected = `{{< video src="/wp-content/uploads/2026/05/my%5Fvideo.mp4" >}}`
	require.Equal(t, expected, replaceVideoShortCode(htmlData))
}

func TestReplaceVideo6(t *testing.T) {
	t.Parallel()
	const htmlData = `<figure class="wp-block-video"><video controls="" src="/wp-content/uploads/2026/05/video.mp4"></video><figcaption class="wp-element-caption">A video example</figcaption></figure>`
	const expected = `{{< video src="/wp-content/uploads/2026/05/video.mp4" >}}`
	require.Equal(t, expected, replaceVideoShortCode(htmlData))
}
