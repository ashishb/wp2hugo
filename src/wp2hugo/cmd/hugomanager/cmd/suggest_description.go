package cmd

import (
	"context"
	"errors"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/descriptionsuggest"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	var colorLogOutput bool
	var hugoDir string
	var updateInline bool
	var limit int

	_suggestDescriptionCmd := &cobra.Command{
		Use:   "suggest-description",
		Short: "Suggests description for all the posts that are missing a description in the front matter",
		Long:  "Suggests description for all the posts that are missing a description in the front matter",
		Run: func(cmd *cobra.Command, args []string) {
			logger.ConfigureLogging(colorLogOutput)
			suggestDescription(cmd.Context(), hugoDir, updateInline, limit)
		},
	}
	_suggestDescriptionCmd.Flags().StringVarP(&hugoDir, "hugo-dir", "", "", "Hugo base directory or any directory containing Hugo markdown files")
	_suggestDescriptionCmd.Flags().BoolVarP(&updateInline, "inline", "i", false, "Add description in markdown files")
	_suggestDescriptionCmd.PersistentFlags().BoolVarP(&colorLogOutput, "color-log-output", "", true,
		"enable colored log output, set false to structured JSON log")
	_suggestDescriptionCmd.Flags().IntVarP(&limit, "limit", "n", 10, "Limit the number of files to update")
	rootCmd.AddCommand(_suggestDescriptionCmd)
}

func suggestDescription(ctx context.Context, hugoDir string, updateInline bool, _limit int) {
	log.Info().Msg("suggest-description command called")

	numHasDescription := 0
	numMissingDescription := 0
	numUpdated := 0
	action := func(path string, updateInline bool) error {
		if numUpdated >= _limit {
			log.Info().
				Int("numUpdated", numUpdated).
				Msg("Limit reached, stopping further updates")
			return nil
		}

		err := descriptionsuggest.ProcessFile(ctx, path, updateInline)
		if err != nil {
			if errors.Is(err, descriptionsuggest.ErrFrontMatterMissingDescription) {
				numMissingDescription++
				return nil
			}
			if errors.Is(err, descriptionsuggest.ErrFrontMatterHasDescription) {
				numHasDescription++
				return nil
			}
			return err
		}

		if updateInline {
			log.Debug().
				Str("path", path).
				Msg("Updated description in front matter")
			numUpdated++
		} else {
			numHasDescription++
		}

		return err
	}
	scanDir(hugoDir, updateInline, action)
	log.Info().
		Int("numHasDescription", numHasDescription).
		Int("numMissingDescription", numMissingDescription).
		Int("numUpdated", numUpdated).
		Msg("suggest-description command completed")
}
