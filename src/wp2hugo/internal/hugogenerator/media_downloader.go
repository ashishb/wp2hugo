package hugogenerator

import (
	"fmt"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"path"
	"strings"
)

func writeFavicon(outputDirPath string, websiteURL string) error {
	log.Debug().Msg("Fetching and writing favicon")
	filePath := path.Join(outputDirPath, "static", "favicon.ico")
	if !strings.HasPrefix(websiteURL, "http") {
		websiteURL = "https://" + websiteURL
	}
	url1 := fmt.Sprintf("%s/favicon.ico", websiteURL)
	return downloadFromURL(url1, filePath)
}

func downloadFromURL(srcURL string, destFilePath string) error {
	log.Debug().
		Str("srcURL", srcURL).
		Str("destFilePath", destFilePath).
		Msg("Downloading from URL")
	if err := createDirIfNotExist(path.Dir(destFilePath)); err != nil {
		return err
	}

	resp, err := http.Get(srcURL)
	if err != nil {
		return fmt.Errorf("error fetching %s: %s", srcURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error fetching %s: %s", srcURL, resp.Status)
	}
	file, err := os.OpenFile(destFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening file %s: %s", file, err)
	}
	defer file.Close()
	return resp.Write(file)
}

func downloadMediaFiles(info wpparser.WebsiteInfo, sourceWebsiteURL string, outputDir string) error {
	log.Warn().Msg("Media files download is not implemented yet")
	return nil
}
