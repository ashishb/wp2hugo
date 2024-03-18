package hugogenerator

import (
	"bytes"
	"fmt"
	"os"
	"path"
)

// setupRssFeedFormat sets up custom guid for RSS feed that is being migrated from WordPress
// to Hugo
func setupRssFeedFormat(siteDir string) error {
	if err := createDirIfNotExist(path.Join(siteDir, "layouts")); err != nil {
		return err
	}
	// Read from the paperMod theme and then generate a new rss.xml that will
	// take precendece over the theme's rss.xml
	data, err := os.ReadFile(path.Join(siteDir, "themes", "PaperMod", "layouts", "_default", "rss.xml"))
	if err != nil {
		return fmt.Errorf("error reading rss.xml from PaperMod theme: %s", err)
	}
	rssFile := path.Join(siteDir, "layouts", "rss.xml")
	return writeFile(rssFile, getModifiedRSSXML(data))
}

func getModifiedRSSXML(data []byte) []byte {
	original := "<guid>{{ .Permalink }}</guid>"
	wordPressCompatible := "" +
		"{{ if .Params.GUID }} " +
		"<guid isPermaLink=\"false\">{{ .Params.guid }}</guid> " +
		"{{ else }} " +
		"<guid isPermaLink=\"false\">{{ .Permalink }}</guid> " +
		"{{ end }}"
	return bytes.ReplaceAll(data, []byte(original), []byte(wordPressCompatible))
}
