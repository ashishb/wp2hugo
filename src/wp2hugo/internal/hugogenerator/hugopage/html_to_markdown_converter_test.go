package hugopage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const _textWithIframe = `
<p><iframe src="https://www.google.com/maps/d/u/0/embed?mid=1lcjyzfxxXcdDP3XkrikfqIJryfFi4ZA&amp;ehbc=2E312F" width="640" height="480"></iframe></p>
`

const _textWithGist = `
<p><a href="https://gist.github.com/lawrencegripper/8e701b0d201e65af0f8bc9b8b0b14207">Gist Link</a></p>
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
