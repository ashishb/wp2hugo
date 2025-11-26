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
		return nil, fmt.Errorf("AudioBitrate field not found in metadata for file: %s", filePath)
	}

	bitRateStr := fmt.Sprintf("%v", bitRateAny)
	bitRateStr = strings.ReplaceAll(bitRateStr, " kbps", "000")
	bitRateStr = strings.ReplaceAll(bitRateStr, " mbps", "000000")
	bitRateInt, err := strconv.Atoi(bitRateStr)
	if err != nil {
		return nil, fmt.Errorf("unable to convert bitrate to int for file %s: %w", filePath, err)
	}

	return &bitRateInt, nil
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

	log.Info().
		Str("inputFile", inputFile).
		Str("outputFile", outputFile).
		Int64("inputFileSize", imageshrinker.GetFileSize(inputFile)).
		Int64("outputFileSize", imageshrinker.GetFileSize(outputFile)).
		Str("bitrate", fmt.Sprintf("%dkbps", bitRate/1000)).
		Msg("Audio bitrate reduction completed successfully")
	return nil
}
