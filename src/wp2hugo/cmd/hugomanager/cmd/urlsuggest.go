package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/urlsuggest"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	HugoDir        string
	UpdateInline   bool
	ColorLogOutput bool
)

var urlSuggestCmd = &cobra.Command{
	Use:   "urlsuggest",
	Short: "Suggests URLs for all the pending/future posts that are missing a URL",
	Long:  "Suggests URLs for all the pending/future posts that are missing a URL",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("URL Suggest command called")
		logger.ConfigureLogging(ColorLogOutput)
		scanDir(HugoDir, UpdateInline)
	},
}

func scanDir(dir string, updateInline bool) {
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
		log.Debug().
			Str("path", path).
			Msg("Processing file")
		_, err = urlsuggest.ProcessFile(path, updateInline)
		if err != nil {
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
