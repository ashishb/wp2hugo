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
		logger.ConfigureLogging(ColorLogOutput)
		action := func(path string, updateInline bool) (*string, error) {
			if Hostname == "" {
				log.Fatal().Msg("Hostname is required")
			}
			return relativelinks.ConvertAbsoluteLinksToRelative(path, updateInline, Hostname)
		}
		scanDir(HugoDir, UpdateInline, action)
	},
}
