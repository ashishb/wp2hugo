package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/imageshrinker"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	var colorLogOutput bool
	var hugoDir string
	var updateInline bool
	var maxSize *int32

	cmd := &cobra.Command{
		Use:   "shrink-images",
		Short: "Shrinks all images to be below a certain width/height",
		Long:  "Shrinks all images to be below a certain width/height",
		Run: func(cmd *cobra.Command, args []string) {
			logger.ConfigureLogging(colorLogOutput)
			shrinkImages(hugoDir, *maxSize, updateInline)
		},
	}

	cmd.Flags().StringVarP(&hugoDir, "hugo-dir", "", "", "Hugo base directory or any directory containing Hugo markdown files")
	cmd.PersistentFlags().BoolVarP(&colorLogOutput, "color-log-output", "", true,
		"enable colored log output, set false to structured JSON log")
	cmd.Flags().BoolVarP(&updateInline, "in-place", "", false, "Shrink images in place")
	maxSize = cmd.PersistentFlags().Int32P("max-dimension", "m", 1500, "Maximum dimension (height/width) of the image, the image would be shrunk proportionately")
	rootCmd.AddCommand(cmd)
}

func shrinkImages(hugoDir string, maxSize int32, updateInline bool) {
	log.Info().Msg("Shrink Images command called")
	if maxSize < 0 {
		log.Fatal().Msg("Maximum dimension cannot be negative")
	}
	action := func(filePath string, updateInline bool) error {
		dims, err := imageshrinker.GetImageDimensions(filePath)
		if err != nil {
			log.Error().
				Err(err).
				Str("filePath", filePath).
				Msg("Error getting image dimensions")
			return err
		}

		if int32(dims.Width()) <= maxSize && int32(dims.Height()) <= maxSize {
			log.Trace().
				Str("filePath", filePath).
				Int("width", dims.Width()).
				Int("height", dims.Height()).
				Msg("Image within size limits, skipping")
			return nil
		}

		log.Debug().
			Str("filePath", filePath).
			Str("size", fmt.Sprintf("%dx%d", dims.Width(), dims.Height())).
			Int32("maxSize", maxSize).
			Msg("Shrinking image")

		if updateInline {
			tmpPath, err := os.CreateTemp("", "shrinked_image_*"+path.Ext(filePath))
			if err != nil {
				log.Error().
					Err(err).
					Str("filePath", filePath).
					Msg("Error creating temporary file for shrunk image")
				return err
			}
			tmpPath.Close()

			if err := imageshrinker.ResizeImage(filePath, tmpPath.Name(), int(maxSize)); err != nil {
				log.Error().
					Err(err).
					Str("filePath", filePath).
					Msg("Error resizing image")
				return err
			}

			if err := os.Rename(tmpPath.Name(), filePath); err != nil {
				log.Error().
					Err(err).
					Str("filePath", filePath).
					Msg("Error replacing original image with shrunk image")
				return err
			}
		}

		return nil
	}
	supportedExtensions := []string{"jpg", "jpeg", "png", "gif", "bmp", "tiff", "webp"}

	scanDir(hugoDir, updateInline, action, supportedExtensions...)
}
