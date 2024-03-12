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
	if err := createDirIfNotExist(path.Join(outputDirPath, "static")); err != nil {
		return err
	}
	filePath := path.Join(outputDirPath, "static", "favicon.ico")
	if !strings.HasPrefix(websiteURL, "http") {
		websiteURL = "https://" + websiteURL
	}
	url1 := fmt.Sprintf("%s/favicon.ico", websiteURL)
	resp, err := http.Get(url1)
	if err != nil {
		return fmt.Errorf("error fetching favicon: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error fetching favicon: %s", resp.Status)
	}
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening favicon file: %s", err)
	}
	defer file.Close()
	return resp.Write(file)
}

func downloadMediaFiles(info wpparser.WebsiteInfo, sourceWebsiteURL string, outputDir string) error {
	log.Warn().Msg("Media files download is not implemented yet")
	return nil
}
