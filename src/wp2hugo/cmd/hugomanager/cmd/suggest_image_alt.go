package cmd

import (
	"context"
	"fmt"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"
	"strings"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/imagealtsuggest"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	var hugoDir string
	var updateInline bool
	var limit int

	suggestImageAltCmd := &cobra.Command{
		Use:   "suggest-image-alt",
		Short: "Suggests image alt text for all the images if missing",
		Long:  "Suggests image alt text for all the images that are missing alt text (useful for accessibility and SEO)",
		Run: func(cmd *cobra.Command, args []string) {
			logger.ConfigureLogging(_colorLogOutput)
			suggestImageAlt(cmd.Context(), hugoDir, updateInline, limit)
		},
	}

	suggestImageAltCmd.Flags().StringVarP(&hugoDir, "hugo-dir", "", "", "Hugo base directory or any directory containing Hugo markdown files")
	suggestImageAltCmd.Flags().BoolVarP(&updateInline, "inline", "i", false, "Add image alt in markdown files")
	suggestImageAltCmd.PersistentFlags().BoolVarP(&_colorLogOutput, "color-log-output", "", true,
		"enable colored log output, set false to structured JSON log")
	suggestImageAltCmd.Flags().IntVarP(&limit, "limit", "n", 10, "Limit the number of images to update")
	rootCmd.AddCommand(suggestImageAltCmd)
}

func suggestImageAlt(ctx context.Context, hugoDir string, updateInline bool, limit int) {
	log.Info().Msg("suggest-image-alt command called")
	numImageWithAlt := 0
	numImageMissingAlt := 0
	numImageUpdated := 0
	action := func(path string, updateInline bool) error {
		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			log.Debug().
				Str("path", path).
				Msg("Skipping non-markdown file")
			return nil
		}
		if numImageUpdated >= limit {
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
	scanDir(hugoDir, updateInline, action)
	log.Info().
		Int("numImageWithAlt", numImageWithAlt).
		Int("numImageMissingAlt", numImageMissingAlt).
		Int("numImageUpdated", numImageUpdated).
		Msg("Image alt text suggestion completed")
}
