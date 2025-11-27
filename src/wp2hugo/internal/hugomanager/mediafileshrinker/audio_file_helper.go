package mediafileshrinker

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/imageshrinker"
	"github.com/barasher/go-exiftool"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

func GetAudioBitrate(filePath string) (*int, error) {
	et, err := exiftool.NewExiftool()
	if err != nil {
		return nil, fmt.Errorf("error when intializing exiftool "+
			"(install exiftool via https://exiftool.org/install.html): %w", err)
	}
	defer et.Close()

	metadatas := et.ExtractMetadata(filePath)
	if len(metadatas) == 0 {
		return nil, fmt.Errorf("no metadata found for file: %s", filePath)
	}

	metadata := metadatas[0]
	if metadata.Err != nil {
		return nil, fmt.Errorf("error reading metadata for file %s: %w", filePath, metadata.Err)
	}

	bitRateAny, ok := metadata.Fields["AudioBitrate"]
	if !ok {
		bitrateAny2, ok2 := metadata.Fields["AvgBitrate"] // For .m4a files
		if ok2 {
			bitRateAny = bitrateAny2
		} else {
			log.Debug().
				Str("filePath", filePath).
				Interface("metadataFields", lo.Keys(metadata.Fields)).
				Msg("Available metadata fields")
			return nil, fmt.Errorf("AudioBitrate field not found in metadata for file: %s", filePath)
		}
	}

	bitRateStr := fmt.Sprintf("%v", bitRateAny)
	bitRateStr = strings.ReplaceAll(bitRateStr, " kbps", "000")
	bitRateStr = strings.ReplaceAll(bitRateStr, " mbps", "000000")
	// Average bit rate is usually in bps, so we need to parse as a float
	bitRateFloat, err := strconv.ParseFloat(bitRateStr, 64)
	if err != nil {
		return nil, fmt.Errorf("unable to convert bitrate to int for file %s: %w", filePath, err)
	}

	return lo.ToPtr(int(bitRateFloat)), nil
}

// ReduceAudioBitrate processes a media file to reduce its audio bitrate
func ReduceAudioBitrate(ctx context.Context, inputFile string, outputFile string, bitRate int) error {
	if bitRate < 1000 {
		return fmt.Errorf("bitrate is %dbps, must be at least 1000 bps", bitRate)
	}

	// assert ffmpeg is installed
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return errors.New("ffmpeg not found in PATH, please install " +
			"ffmpeg (https://www.ffmpeg.org/download.html) to use this feature")
	}

	//nolint:gosec  // it is a false positive
	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-i", inputFile,
		"-y",                                     // Overwrite output file if it exists
		"-b:a", fmt.Sprintf("%dk", bitRate/1000), // Target audio bitrate e.g. 48k
		outputFile,
	)

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Include FFmpeg's output in the error for easier debugging
		return fmt.Errorf("FFmpeg failed: %w\nOutput: %s", err, output)
	}

	originalSize := imageshrinker.GetFileSize(inputFile)
	newSize := imageshrinker.GetFileSize(outputFile)
	shrunkPct := 100.0 * (float64(originalSize-newSize) / float64(originalSize))

	log.Info().
		Str("inputFile", inputFile).
		Str("outputFile", outputFile).
		Int64("inputFileSize", originalSize).
		Int64("outputFileSize", newSize).
		Str("shrinkBy", fmt.Sprintf("%.0f%%", shrunkPct)).
		Str("bitrate", fmt.Sprintf("%dkbps", bitRate/1000)).
		Msg("Audio bitrate reduction completed successfully")
	return nil
}
