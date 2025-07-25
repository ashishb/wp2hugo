package cmd

import (
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/urlsuggest"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var _colorLogOutput bool

func init() {
	var hugoDir string
	var updateInline bool

	urlSuggestCmd := &cobra.Command{
		Use:   "suggest-url",
		Short: "Suggests URLs for all the pending/future posts that are missing a URL",
		Long:  "Suggests URLs for all the pending/future posts that are missing a URL",
		Run: func(cmd *cobra.Command, args []string) {
			logger.ConfigureLogging(_colorLogOutput)
			suggestURL(hugoDir, updateInline)
		},
	}

	urlSuggestCmd.Flags().StringVarP(&hugoDir, "hugo-dir", "", "", "Hugo base directory or any directory containing Hugo markdown files")
	urlSuggestCmd.Flags().BoolVarP(&updateInline, "in-place", "", false, "Update URLs in markdown files")
	urlSuggestCmd.PersistentFlags().BoolVarP(&_colorLogOutput, "color-log-output", "", true,
		"enable colored log output, set false to structured JSON log")
	rootCmd.AddCommand(urlSuggestCmd)
}

func suggestURL(hugoDir string, updateInline bool) {
	log.Info().Msg("URL Suggest command called")
	action := func(path string, updateInline bool) error {
		_, err := urlsuggest.ProcessFile(path, updateInline)
		return err
	}
	scanDir(hugoDir, updateInline, action)
}
