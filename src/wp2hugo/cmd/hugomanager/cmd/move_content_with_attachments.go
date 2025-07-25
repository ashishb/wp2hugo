package cmd

import (
	"context"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/contentmigratorv1"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	var colorLogOutput bool
	var hugoDir string
	var updateInline bool

	moveContentWithAttachmentsCmd := &cobra.Command{
		Use:   "move-post-next-to-attachments",
		Short: "Move markdown blog posts with attachments to a single directory",
		Long: "Update the posts with attachments to a single directory where the attachments are stored next " +
			"to the markdown file for better long-term maintenance",
		Run: func(cmd *cobra.Command, args []string) {
			logger.ConfigureLogging(colorLogOutput)
			moveContentNextToAttachments(cmd.Context(), hugoDir, updateInline)
		},
	}

	moveContentWithAttachmentsCmd.Flags().StringVarP(&hugoDir, "hugo-dir", "", "", "Hugo base directory or any directory containing Hugo markdown files")
	moveContentWithAttachmentsCmd.PersistentFlags().BoolVarP(&colorLogOutput, "color-log-output", "", true,
		"enable colored log output, set false to structured JSON log")
	moveContentWithAttachmentsCmd.Flags().BoolVarP(&updateInline, "in-place", "", false, "Update titles in in markdown files")
	rootCmd.AddCommand(moveContentWithAttachmentsCmd)
}

func moveContentNextToAttachments(ctx context.Context, hugoDir string, updateInline bool) {
	log.Info().Msg("Move content with attachments command called")

	modifiedCount := 0
	action := func(path string, updateInline bool) error {
		processedFile, err := contentmigratorv1.ProcessFile(ctx, path, updateInline)
		if err != nil {
			return err
		}
		if *processedFile {
			modifiedCount++
		}
		log.Debug().
			Str("path", path).
			Bool("updateInline", updateInline).
			Int("modifiedCount", modifiedCount).
			Msg("Processed file")
		return nil
	}
	scanDir(hugoDir, updateInline, action)
}
