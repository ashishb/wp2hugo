package cmd

import (
	"fmt"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/imagealtsuggest"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strings"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"
)

func init() {
	_suggestImageAltCmd.Flags().StringVarP(&_hugoDir, "hugo-dir", "", "", "Hugo base directory or any directory containing Hugo markdown files")
	_suggestImageAltCmd.Flags().BoolVarP(&_updateInline, "inline", "i", false, "Add image alt in markdown files")
	_suggestImageAltCmd.PersistentFlags().BoolVarP(&_colorLogOutput, "color-log-output", "", true,
		"enable colored log output, set false to structured JSON log")
	_suggestImageAltCmd.Flags().IntVarP(&_limit, "limit", "n", 10, "Limit the number of images to update")
	rootCmd.AddCommand(_suggestImageAltCmd)
}

var _suggestImageAltCmd = &cobra.Command{
	Use:   "suggest-image-alt",
	Short: "Suggests image alt text for all the images if missing",
	Long:  "Suggests image alt text for all the images that are missing alt text (useful for accessibility and SEO)",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("suggest-image-alt command called")
		logger.ConfigureLogging(_colorLogOutput)
		numImageWithAlt := 0
		numImageMissingAlt := 0
		numImageUpdated := 0
		ctx := cmd.Context()
		action := func(path string, updateInline bool) error {
			if !strings.HasSuffix(strings.ToLower(path), ".md") {
				log.Debug().
					Str("path", path).
					Msg("Skipping non-markdown file")
				return nil
			}
			if numImageUpdated >= _limit {
				log.Info().
					Int("numImageUpdated", numImageUpdated).
					Msg("Limit reached, stopping further updates")
				return nil
			}

			result, err := imagealtsuggest.ProcessFile(ctx, path, updateInline)
			if err != nil {
				return fmt.Errorf("failed to process file %s: %w", path, err)
			}

			numImageWithAlt += result.NumImageWithAlt()
			numImageMissingAlt += result.NumImageMissingAlt()
			numImageUpdated += result.NumImageUpdated()
			if result.NumImageUpdated() > 0 {
				log.Debug().
					Str("path", path).
					Int("numImageUpdated", result.NumImageUpdated()).
					Msg("Updated description in front matter")
			}

			return err
		}
		scanDir(_hugoDir, _updateInline, action)
		log.Info().
			Int("numImageWithAlt", numImageWithAlt).
			Int("numImageMissingAlt", numImageMissingAlt).
			Int("numImageUpdated", numImageUpdated).
			Msg("Image alt text suggestion completed")
	},
}
