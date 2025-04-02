package cmd

import (
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/urlsuggest"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	_hugoDir        string
	_updateInline   bool
	_colorLogOutput bool
)

func init() {
	urlSuggestCmd.Flags().StringVarP(&_hugoDir, "hugo-dir", "", "", "Hugo base directory or any directory containing Hugo markdown files")
	urlSuggestCmd.Flags().BoolVarP(&_updateInline, "in-place", "", false, "Update titles in in markdown files")
	urlSuggestCmd.PersistentFlags().BoolVarP(&_colorLogOutput, "color-log-output", "", true,
		"enable colored log output, set false to structured JSON log")
	rootCmd.AddCommand(urlSuggestCmd)

}

var urlSuggestCmd = &cobra.Command{
	Use:   "suggest-url",
	Short: "Suggests URLs for all the pending/future posts that are missing a URL",
	Long:  "Suggests URLs for all the pending/future posts that are missing a URL",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("URL Suggest command called")
		logger.ConfigureLogging(_colorLogOutput)
		action := func(path string, updateInline bool) error {
			_, err := urlsuggest.ProcessFile(path, updateInline)
			return err
		}
		scanDir(_hugoDir, _updateInline, action)
	},
}
