package cmd

import (
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/relativelinks"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	var hostname string
	var hugoDir string
	var updateInline bool

	relativeLinksCmd := &cobra.Command{
		Use:   "make-absolute-internal-links-relative",
		Short: "Converts all the absolute internal links to relative links",
		Long:  "Converts all the absolute internal links to relative links",
		Run: func(cmd *cobra.Command, args []string) {
			moveAbsoluteLinksToRelative(hugoDir, updateInline, hostname)
		},
	}

	relativeLinksCmd.Flags().StringVarP(&hugoDir, "hugo-dir", "", "", "Hugo base directory or any directory containing Hugo markdown files")
	relativeLinksCmd.Flags().BoolVarP(&updateInline, "in-place", "", false, "Update URLs in in markdown files")
	relativeLinksCmd.PersistentFlags().BoolVarP(&_colorLogOutput, "color-log-output", "", true,
		"enable colored log output, set false to structured JSON log")
	relativeLinksCmd.Flags().StringVarP(&hostname, "hostname", "", "", "All hostname under this will be considered internal links")
	rootCmd.AddCommand(relativeLinksCmd)
}

func moveAbsoluteLinksToRelative(hugoDir string, updateInline bool, hostname string) {
	log.Info().Msg("Relative Links command called")
	logger.ConfigureLogging(_colorLogOutput)
	if hostname == "" {
		log.Fatal().Msg("hostname is required")
	}
	action := func(path string, updateInline bool) error {
		return relativelinks.ConvertAbsoluteLinksToRelative(path, updateInline, hostname)
	}
	scanDir(hugoDir, updateInline, action)
}
