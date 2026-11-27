package hugopage

import (
	"testing"
)

func TestEmbedConversion(t *testing.T) {
	t.Parallel()
	// Test case name -> (input HTML, expected Markdown)
	testCases := map[string]struct {
		html     string
		expected string
	}{
		"Gutenberg Vimeo embed block": {
			html: `<!-- wp:embed {"url":"https://vimeo.com/76979871","type":"video","providerNameSlug":"vimeo","responsive":true} -->
<figure class="wp-block-embed is-type-video is-provider-vimeo wp-block-embed-vimeo"><div class="wp-block-embed__wrapper">
https://vimeo.com/76979871
</div></figure>
<!-- /wp:embed -->`,
			expected: "{{< vimeo 76979871 >}}",
		},
		"Vimeo URL on its own line with query parameters": {
			html:     "<p>Intro</p>\n<p>https://vimeo.com/76979871?share=copy</p>\n<p>End</p>",
			expected: "Intro\n\n{{< vimeo 76979871 >}}\n\nEnd",
		},
		"Vimeo URL inside a sentence is a link, not an embed": {
			html:     `<p>See https://vimeo.com/76979871 for more</p>`,
			expected: "See https://vimeo.com/76979871 for more",
		},
		"unlisted Vimeo URL is kept, since the shortcode does not support the hash": {
			html:     `<p>https://vimeo.com/76979871/0123abcdef</p>`,
			expected: "https://vimeo.com/76979871/0123abcdef",
		},
		"Vimeo URL in the embed shortcode": {
			html:     "<p>Intro</p>\n[embed]https://vimeo.com/76979871[/embed]\n<p>End</p>",
			expected: "Intro\n\n{{< vimeo 76979871 >}}\n\nEnd",
		},
		"YouTube URL in the embed shortcode": {
			html:     `<p>[embed]https://www.youtube.com/watch?v=gJ7AAJXHeeg[/embed]</p>`,
			expected: "{{< youtube gJ7AAJXHeeg >}}",
		},
		"embed shortcode with attributes": {
			html:     `<p>[embed width="500" height="300"]https://www.youtube.com/watch?v=gJ7AAJXHeeg[/embed]</p>`,
			expected: "{{< youtube gJ7AAJXHeeg >}}",
		},
		"embed shortcode with an unsupported provider keeps the URL": {
			html:     `<p>[embed]https://www.dailymotion.com/video/x8abcde[/embed]</p>`,
			expected: "https://www.dailymotion.com/video/x8abcde",
		},
		"Gutenberg Instagram embed block with underscore in the post ID": {
			html: `<!-- wp:embed {"url":"https://www.instagram.com/p/Cx_OWiQNP2MO/","type":"rich","providerNameSlug":"instagram"} -->
<figure class="wp-block-embed is-type-rich is-provider-instagram wp-block-embed-instagram"><div class="wp-block-embed__wrapper">
https://www.instagram.com/p/Cx_OWiQNP2MO/
</div></figure>
<!-- /wp:embed -->`,
			expected: "{{< instagram Cx_OWiQNP2MO >}}",
		},
		"Instagram URL on its own line with query parameters": {
			html:     `<p>https://www.instagram.com/p/CxOWiQNP2MO/?utm_source=ig_web_copy_link</p>`,
			expected: "{{< instagram CxOWiQNP2MO >}}",
		},
		"Instagram profile URL is kept": {
			html:     `<p>https://www.instagram.com/someone/</p>`,
			expected: "https://www.instagram.com/someone/",
		},
		"Vimeo iframe": {
			html:     `<p><iframe src="https://player.vimeo.com/video/76979871?title=0&amp;byline=0" width="640" height="360" frameborder="0" allowfullscreen></iframe></p>`,
			expected: "{{< vimeo 76979871 >}}",
		},
		"unlisted Vimeo iframe is kept as HTML": {
			html:     `<p><iframe src="https://player.vimeo.com/video/76979871?h=0123abcdef" width="640" height="360"></iframe></p>`,
			expected: `<iframe src="https://player.vimeo.com/video/76979871?h=0123abcdef" width="640" height="360"></iframe>`,
		},
		"other iframes are kept as HTML instead of being dropped": {
			html:     `<p>Before</p><p><iframe src="https://www.dailymotion.com/embed/video/x8abcde" width="640" height="360" allowfullscreen></iframe></p><p>After</p>`,
			expected: "Before\n\n<iframe src=\"https://www.dailymotion.com/embed/video/x8abcde\" width=\"640\" height=\"360\" allowfullscreen=\"\"></iframe>\n\nAfter",
		},
		"YouTube iframe is still converted": {
			html:     `<p><iframe width="560" height="315" src="https://www.youtube.com/embed/Wz6ml5SpkKM"></iframe></p>`,
			expected: `{{< youtube id="Wz6ml5SpkKM" >}}`,
		},
		"Google Maps iframe is still converted": {
			html:     `<p><iframe src="https://www.google.com/maps/d/u/0/embed?mid=1lcjyzfxxXcdDP3XkrikfqIJryfFi4ZA" width="640" height="480"></iframe></p>`,
			expected: `{{< googlemaps src="1lcjyzfxxXcdDP3XkrikfqIJryfFi4ZA" width=640 height=480 >}}`,
		},
		"URLs in code blocks are not converted": {
			html:     "<pre><code>https://vimeo.com/76979871\nhttps://www.instagram.com/p/CxOWiQNP2MO/\n[embed]https://vimeo.com/1[/embed]</code></pre>",
			expected: "```\nhttps://vimeo.com/76979871\nhttps://www.instagram.com/p/CxOWiQNP2MO/\n[embed]https://vimeo.com/1[/embed]\n```",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			testMarkdownExtractor(t, tc.html, tc.expected)
		})
	}
}
