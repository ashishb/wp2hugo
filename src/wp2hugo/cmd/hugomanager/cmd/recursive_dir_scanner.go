package cmd

import (
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
)

func scanDir(dir string, updateInline bool, action func(string, bool) (*string, error)) {
	if dir == "" {
		log.Fatal().Msg("Hugo directory not provided")
	}
	log.Info().
		Str("dir", dir).
		Msg("Scanning directory")
	if !utils.DirExists(dir) {
		log.Fatal().
			Str("dir", dir).
			Msg("Directory does not exist")
	}
	failed := 0
	processed := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error().
				Err(err).
				Str("path", path).
				Msg("Error walking directory")
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		log.Trace().
			Str("path", path).
			Msg("Processing file")
		_, err = action(path, updateInline)
		if err != nil {
			log.Warn().
				Err(err).
				Str("path", path).
				Msg("Error processing file")
			failed++
		} else {
			processed++
		}
		return nil
	})
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Error processing directory")
	}
	log.Info().
		Int("processed", processed).
		Int("failed", failed).
		Msg("Finished processing")
}
