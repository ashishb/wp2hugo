package hugogenerator

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"regexp"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/rs/zerolog/log"
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
			return fmt.Errorf("error unescaping filename %s: %w", fileName, err)
		}
		log.Info().
			Str("fileName", fileName).
			Str("newFileName", tmp1).
			Msg("Unescaped filename")
		fileName = tmp1
		destFilePath = path.Join(path.Dir(destFilePath), fileName)
	}

	file, err := os.OpenFile(destFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", destFilePath, err)
	}

	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("error writing to file %s: %w", destFilePath, err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("error closing file %s: %w", destFilePath, err)
	}

	return err
}
