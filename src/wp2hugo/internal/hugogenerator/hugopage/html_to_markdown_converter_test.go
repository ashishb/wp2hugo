package hugopage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const _textWithIframe = `
<p><iframe src="https://www.google.com/maps/d/u/0/embed?mid=1lcjyzfxxXcdDP3XkrikfqIJryfFi4ZA&amp;ehbc=2E312F" width="640" height="480"></iframe></p>
`

const _textWithGist = `
<figure class="wp-block-embed is-type-rich is-provider-embed-handler wp-block-embed-embed-handler"><div class="wp-block-embed__wrapper">
https://gist.github.com/lawrencegripper/8e701b0d201e65af0f8bc9b8b0b14207
</div></figure>
`

const _textMarkdownGist = `
<body>
some text
\[gist https://gist.github.com/lawrencegripper/6bee7de123bea1936359\]
some more text
</body>
`

func TestIframe(t *testing.T) {
	converter := getMarkdownConverter()
	result, err := converter.ConvertString(_textWithIframe)
	assert.NoError(t, err)
	assert.Contains(t, result, `{{< googlemaps src="1lcjyzfxxXcdDP3XkrikfqIJryfFi4ZA" width=640 height=480 >}}`)
}

func TestGist(t *testing.T) {
	converter := getMarkdownConverter()
	result, err := converter.ConvertString(_textWithGist)
	assert.NoError(t, err)
	assert.Contains(t, result, `{{< gist lawrencegripper 8e701b0d201e65af0f8bc9b8b0b14207 >}}`)
}

func TestMarkdownGist(t *testing.T) {
	converter := getMarkdownConverter()
	result, err := converter.ConvertString(_textMarkdownGist)
	assert.NoError(t, err)
	assert.Contains(t, result, `{{< gist lawrencegripper 6bee7de123bea1936359 >}}`)
}
