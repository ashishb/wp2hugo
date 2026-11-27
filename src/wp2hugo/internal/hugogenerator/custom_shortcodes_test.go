package hugogenerator

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

// Hugo removed its built-in "gist" shortcode in v0.156.0, so the generated site must ship its own
func TestWriteCustomShortCodesIncludesGist(t *testing.T) {
	t.Parallel()
	siteDir := t.TempDir()
	require.NoError(t, WriteCustomShortCodes(siteDir))

	content, err := os.ReadFile(path.Join(siteDir, "layouts", "shortcodes", "gist.html"))
	require.NoError(t, err)
	require.Contains(t, string(content), `<script src="https://gist.github.com/{{ index .Params 0 }}/{{ index .Params 1 }}.js`)
}
