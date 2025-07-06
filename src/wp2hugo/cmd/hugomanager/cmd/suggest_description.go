package cmd

import (
	"errors"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/descriptionsuggest"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var _limit int

func init() {
	_suggestDescriptionCmd.Flags().StringVarP(&_hugoDir, "hugo-dir", "", "", "Hugo base directory or any directory containing Hugo markdown files")
	_suggestDescriptionCmd.Flags().BoolVarP(&_updateInline, "inline", "i", false, "Add description in markdown files")
	_suggestDescriptionCmd.PersistentFlags().BoolVarP(&_colorLogOutput, "color-log-output", "", true,
		"enable colored log output, set false to structured JSON log")
	_suggestDescriptionCmd.Flags().IntVarP(&_limit, "limit", "n", 10, "Limit the number of files to update")
	rootCmd.AddCommand(_suggestDescriptionCmd)
}

var _suggestDescriptionCmd = &cobra.Command{
	Use:   "suggest-description",
	Short: "Suggests description for all the posts that are missing a description in the front matter",
	Long:  "Suggests description for all the posts that are missing a description in the front matter",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("suggest-description command called")
		logger.ConfigureLogging(_colorLogOutput)

		numHasDescription := 0
		numMissingDescription := 0
		numUpdated := 0
		ctx := cmd.Context()
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
		scanDir(_hugoDir, _updateInline, action)
		log.Info().
			Int("numHasDescription", numHasDescription).
			Int("numMissingDescription", numMissingDescription).
			Int("numUpdated", numUpdated).
			Msg("suggest-description command completed")
	},
}
