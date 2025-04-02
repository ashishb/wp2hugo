package cmd

import (
	"errors"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/descriptionsuggest"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"
)

func init() {
	_suggestDescriptionCmd.Flags().StringVarP(&_hugoDir, "hugo-dir", "", "", "Hugo base directory or any directory containing Hugo markdown files")
	_suggestDescriptionCmd.Flags().BoolVarP(&_updateInline, "in-place", "", false, "Add description in markdown files")
	_suggestDescriptionCmd.PersistentFlags().BoolVarP(&_colorLogOutput, "color-log-output", "", true,
		"enable colored log output, set false to structured JSON log")
	rootCmd.AddCommand(_suggestDescriptionCmd)

}

var _suggestDescriptionCmd = &cobra.Command{
	Use:   "suggest-description",
	Short: "Suggests description for all the posts that are missing a description in the front matter",
	Long:  "Suggests description for all the posts that are missing a description in the front matter",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("URL Suggest command called")
		logger.ConfigureLogging(_colorLogOutput)

		numHasDescription := 0
		numMissingDescription := 0
		action := func(path string, updateInline bool) error {
			err := descriptionsuggest.ProcessFile(path, updateInline)
			if err == nil {
				numHasDescription++
				return nil
			}

			if errors.Is(err, descriptionsuggest.ErrFrontMatterMissingDescription) {
				numMissingDescription++
				return nil
			}

			return err
		}
		scanDir(_hugoDir, _updateInline, action)
		log.Info().
			Int("numHasDescription", numHasDescription).
			Int("numMissingDescription", numMissingDescription).
			Msg("suggest-description command completed")
	},
}
