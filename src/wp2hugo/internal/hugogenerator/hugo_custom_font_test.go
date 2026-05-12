package hugogenerator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetupFontUsesPrimaryHeadFileWhenAvailable(t *testing.T) {
	t.Parallel()
	siteDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(siteDir, "themes/PaperMod/layouts/partials"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(siteDir, "themes/PaperMod/assets/css/extended"), 0o755))

	require.NoError(t, setupFont(siteDir, "Roboto"))

	primaryHeadFile := filepath.Join(siteDir, _outputHeadFile)
	_, err := os.Stat(primaryHeadFile)
	require.NoError(t, err)

	data, err := os.ReadFile(primaryHeadFile)
	require.NoError(t, err)
	require.Contains(t, string(data), "Roboto")
}

func TestSetupFontFallsBackToUnderscorePartialsPath(t *testing.T) {
	t.Parallel()
	siteDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(siteDir, "themes/PaperMod/layouts/_partials"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(siteDir, "themes/PaperMod/assets/css/extended"), 0o755))

	require.NoError(t, setupFont(siteDir, "Inter"))

	fallbackHeadFile := filepath.Join(siteDir, _outputHeadFileFallback)
	_, err := os.Stat(fallbackHeadFile)
	require.NoError(t, err)

	data, err := os.ReadFile(fallbackHeadFile)
	require.NoError(t, err)
	require.Contains(t, string(data), "Inter")
}
