package hugogenerator

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
)

const _extendedHeaderData = `
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=%s&display=swap" rel="stylesheet">
`

const _outputHeadFile = "themes/PaperMod/layouts/partials/extend_head.html"

const _customFontCSS = `
body {
	font-family: '%s', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
	font-size: 18px;
	line-height: 1.6;
	word-break: break-word;
	background: var(--theme);
}
`

const _customCSS = `
.gallery {
	display: flex;
	flex-wrap: wrap;
}
.gallery figure,
.gallery figure img {
	text-align: center;
}
.gallery figure img {
	margin: 1rem auto;
}
.gallery-cols-1 figure {
	width: 100%;
}
.gallery-cols-2 figure {
	width: 50%;
}
.gallery-cols-3 figure {
	width: 33.3333333333%;
}
.gallery-cols-4 figure {
	width: 25%;
}
.gallery-cols-5 figure {
	width: 25%;
}
.gallery-cols-6 figure {
	width: 16.666666666%;
}
`

const _outputCssFile = "themes/PaperMod/assets/css/extended/blank.css"

// Custom font for Hugo's papermod theme
// Ref: https://forum.wildserver.ru/viewtopic.php?t=18
func setupFont(siteDir string, fontName string) error {
	err1 := appendFile(filepath.Join(siteDir, _outputHeadFile), fmt.Sprintf(_extendedHeaderData, fontName))
	err2 := appendFile(filepath.Join(siteDir, _outputCssFile), fmt.Sprintf(_customFontCSS, fontName))
	err3 := appendFile(filepath.Join(siteDir, _outputCssFile), _customCSS)
	return errors.Join(err1, err2, err3)
}

func appendFile(outputFilePath string, data string) error {
	log.Info().
		Str("location", outputFilePath).
		Msgf("Writing custom font to %s", outputFilePath)
	f, err := os.OpenFile(outputFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(data); err != nil {
		return err
	}
	return nil
}
