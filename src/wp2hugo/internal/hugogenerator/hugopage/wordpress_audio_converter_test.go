package hugopage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceAudio1(t *testing.T) {
	const htmlData = `[audio src="/wp-content/uploads/sites/3/2020/07/session_2020-07-02.mp3"][/audio]`
	const expected = `{{< audio src="/wp-content/uploads/sites/3/2020/07/session%5F2020-07-02.mp3" >}}`
	assert.Equal(t, expected, replaceAudioShortCode(htmlData))
}

func TestReplaceAudio2(t *testing.T) {
	const htmlData = `[audio mp3="/wp-content/uploads/sites/3/2020/07/session_2020-07-02.mp3"]`
	const expected = `{{< audio src="/wp-content/uploads/sites/3/2020/07/session%5F2020-07-02.mp3" >}}`
	assert.Equal(t, expected, replaceAudioShortCode(htmlData))
}

func TestReplaceAudio3(t *testing.T) {
	const htmlData = `[audio m4a="/wp-content/uploads/sites/3/2020/07/session_2020-07-02.m4a" mp3="/wp-content/uploads/sites/3/2020/07/session_2020-07-02.mp3"]`
	const expected = `{{< audio src="/wp-content/uploads/sites/3/2020/07/session%5F2020-07-02.m4a" >}}`
	assert.Equal(t, expected, replaceAudioShortCode(htmlData))
}
