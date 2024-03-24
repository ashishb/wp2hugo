package hugopage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPreTagExtractor2(t *testing.T) {
	const example1 = `<pre class="lang:js decode:true">document.querySelector("video").playbackRate = 2.0;   // For 2X speed-up</pre>`
	const example2 = `<pre class="theme:solarized-dark lang:sh decode:true">echo "whatever"</pre>`
	const example3 = `<pre class="lang:sh decode:true"># Sample invocation:\n</pre>`
	assert.True(t, _preTagExtractor2.MatchString(example1))
	assert.True(t, _preTagExtractor2.MatchString(example2))
	assert.True(t, _preTagExtractor2.MatchString(example3))

	result3 := _preTagExtractor2.FindAllStringSubmatch(example3, -1)
	assert.Equal(t, 1, len(result3))
	assert.Equal(t, 3, len(result3[0]))
	assert.Equal(t, "sh", result3[0][1])
}
