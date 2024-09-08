package cmd

import (
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/sitesummary"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	siteSummaryCmd = &cobra.Command{
		Use:   "sitesummary",
		Short: "Print site stats (e.g. number of posts, number of drafts etc.)",
		Long:  "Print site stats (e.g. number of posts, number of drafts etc.)",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info().Msg("Site Summary command called")
			logger.ConfigureLogging(ColorLogOutput)
			generateDirSummary(HugoDir)
		},
	}
)

func generateDirSummary(dir string) {
	if dir == "" {
		log.Fatal().Msg("Hugo directory not provided")
	}
	log.Info().
		Str("dir", dir).
		Msg("Scanning directory")
	if !utils.DirExists(dir) {
		log.Fatal().
			Str("dir", dir).
			Msg("Directory does not exist")
	}
	summary, err := sitesummary.ScanDir(dir)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Error scanning directory")
	}
	log.Info().
		Int("posts", summary.Posts()).
		Int("drafts", summary.Drafts()).
		Int("future", summary.Future()).
		Msg("Site Summary")

	for _, draft := range summary.DraftPostPaths(10) {
		log.Info().
			Str("Path", draft.Path).
			Msg("Draft post")
	}

	for _, future := range summary.FuturePostPaths(10) {
		log.Info().
			Str("Path", future.Path).
			Str("TimeLeftToPublish", future.RelativeTime()).
			Msg("Future post")
	}
}
