package cmd

import (
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/relativelinks"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var Hostname string

var relativeLinksCmd = &cobra.Command{
	Use:   "make-absolute-internal-links-relative",
	Short: "Converts all the absolute internal links to relative links",
	Long:  "Converts all the absolute internal links to relative links",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("Relative Links command called")
		logger.ConfigureLogging(_colorLogOutput)
		action := func(path string, updateInline bool) (*string, error) {
			if Hostname == "" {
				log.Fatal().Msg("Hostname is required")
			}
			return relativelinks.ConvertAbsoluteLinksToRelative(path, updateInline, Hostname)
		}
		scanDir(_hugoDir, _updateInline, action)
	},
}

func init() {
	relativeLinksCmd.Flags().StringVarP(&_hugoDir, "hugo-dir", "", "", "Hugo base directory or any directory containing Hugo markdown files")
	relativeLinksCmd.Flags().BoolVarP(&_updateInline, "in-place", "", false, "Update titles in in markdown files")
	relativeLinksCmd.PersistentFlags().BoolVarP(&_colorLogOutput, "color-log-output", "", true,
		"enable colored log output, set false to structured JSON log")
	relativeLinksCmd.Flags().StringVarP(&Hostname, "hostname", "", "", "All hostname under this will be considered internal links")
	rootCmd.AddCommand(relativeLinksCmd)
}
