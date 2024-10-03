package hugogenerator

import (
	"fmt"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/rs/zerolog/log"
	"io"
	"net/url"
	"os"
	"path"
	"regexp"
)

func writeFavicon(outputDirPath string, faviconData io.Reader) error {
	log.Debug().Msg("Writing favicon")
	return download(path.Join(outputDirPath, "favicon.ico"), faviconData)
}

// Match %dd
var _hexPattern = regexp.MustCompile(`%[0-9a-fA-F]{2}`)

func download(destFilePath string, reader io.Reader) error {
	log.Debug().
		Str("destFilePath", destFilePath).
		Msg("Downloading from URL")
	if err := utils.CreateDirIfNotExist(path.Dir(destFilePath)); err != nil {
		return err
	}
	fileName := path.Base(destFilePath)
	if _hexPattern.MatchString(fileName) {
		tmp1, err := url.PathUnescape(fileName)
		if err != nil {
			return fmt.Errorf("error unescaping filename %s: %s", fileName, err)
		}
		log.Info().
			Str("fileName", fileName).
			Str("newFileName", tmp1).
			Msg("Unescaped filename")
		fileName = tmp1
		destFilePath = path.Join(path.Dir(destFilePath), fileName)
	}

	file, err := os.OpenFile(destFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening file %s: %s", destFilePath, err)
	}
	defer file.Close()
	_, err = io.Copy(file, reader)
	return err
}
