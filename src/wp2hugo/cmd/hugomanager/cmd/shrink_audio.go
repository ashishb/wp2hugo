package cmd

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/mediafileshrinker"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	var colorLogOutput bool
	var hugoDir string
	var updateInline bool
	var maxSize *int

	cmd := &cobra.Command{
		Use:   "shrink-audio-files",
		Short: "Shrinks all audio files to be below a certain bitrate",
		Long:  "Shrinks all audio files to be below a certain bitrate",
		Run: func(cmd *cobra.Command, args []string) {
			logger.ConfigureLogging(colorLogOutput)
			shrinkAudioFiles(cmd.Context(), hugoDir, *maxSize, updateInline)
		},
	}

	cmd.Flags().StringVarP(&hugoDir, "hugo-dir", "d", "", "Hugo base directory or any directory containing Hugo markdown files")
	cmd.PersistentFlags().BoolVarP(&colorLogOutput, "color-log-output", "", true,
		"enable colored log output, set false to structured JSON log")
	cmd.Flags().BoolVarP(&updateInline, "in-place", "i", false, "Shrink audio files in place")
	maxSize = cmd.PersistentFlags().IntP("max-bitrate", "m", 48_000, "Maximum bitrate of the audio files")
	rootCmd.AddCommand(cmd)
}

func shrinkAudioFiles(ctx context.Context, hugoDir string, maxBitRate int, updateInline bool) {
	log.Info().Msg("Shrink audio files command called")
	if maxBitRate < 0 {
		log.Fatal().Msg("Maximum dimension cannot be negative")
	}
	action := func(filePath string, updateInline bool) error {
		bitrate, err := mediafileshrinker.GetAudioBitrate(filePath)
		if err != nil {
			log.Error().
				Err(err).
				Str("filePath", filePath).
				Msg("Error getting audio file bitrate")
			return err
		}

		if *bitrate <= maxBitRate {
			log.Trace().
				Str("filePath", filePath).
				Int("bitrate", *bitrate).
				Msg("Audio file within bitrate limits, skipping")
			return nil
		}

		log.Debug().
			Str("filePath", filePath).
			Str("bitrate", fmt.Sprintf("%dkbps", *bitrate/1000)).
			Str("maxBitRate", fmt.Sprintf("%dkbps", maxBitRate/1000)).
			Msg("Shrinking audio file to be within bitrate limits")

		if updateInline {
			tmpPath, err := os.CreateTemp("", "shrinked_media_*"+path.Ext(filePath))
			if err != nil {
				log.Error().
					Err(err).
					Str("filePath", filePath).
					Msg("Error creating temporary file for shrunk audio file")
				return err
			}
			tmpPath.Close()

			if err := mediafileshrinker.ReduceAudioBitrate(ctx, filePath, tmpPath.Name(), maxBitRate); err != nil {
				log.Error().
					Err(err).
					Str("filePath", filePath).
					Msg("Error resizing audio file")
				return err
			}

			if err := os.Rename(tmpPath.Name(), filePath); err != nil {
				log.Error().
					Err(err).
					Str("filePath", filePath).
					Msg("Error replacing original audio file with shrunk audio file")
				return err
			}
		}

		return nil
	}
	supportedExtensions := []string{"mp3", "wav", "aac", "flac", "ogg", "m4a"}

	scanDir(hugoDir, updateInline, action, supportedExtensions...)
}
