package main

import (
	"flag"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugogenerator"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/mediacache"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"strings"
)

var (
	sourceFile                     = flag.String("source", "", "file path to the source WordPress XML file")
	outputDir                      = flag.String("output", "/tmp", "dir path to the write the Hugo generated data to")
	downloadMedia                  = flag.Bool("download-media", false, "download media files embedded in the WordPress content")
	continueOnMediaDownloadFailure = flag.Bool("continue-on-media-download-error", false, "continue processing even if one more more media download fails")
	authors                        = flag.String("authors", "", "CSV list of author name(s), if provided, only posts by these authors will be processed")
	// This is useful for repeated executions of the tool to avoid downloading the media files again
	// Mostly for development and not for the production use
	mediaCacheDir = flag.String("media-cache-dir", path.Join("/tmp/wp2hugo-cache"), "dir path to cache the downloaded media files")
	// Custom font for Hugo's papermod theme
	font           = flag.String("font", "Lexend", "custom font for the output website")
	colorLogOutput = flag.Bool("color-log-output", true, "enable colored log output, set false to structured JSON log")
)

func main() {
	flag.Parse()

	// Set log level
	logger.ConfigureLogging(*colorLogOutput)
	if len(*sourceFile) == 0 {
		log.Fatal().Msg("Source file is required")
	}
	if len(*outputDir) == 0 {
		log.Fatal().Msg("Output directory is required")
	}
	err := handle(*sourceFile)
	if err != nil {
		log.Fatal().Msgf("Error: %s", err)
	}
}

func handle(filePath string) error {
	log.Debug().
		Str("source", filePath).
		Msg("Reading website export")
	websiteInfo, err := getWebsiteInfo(filePath)
	if err != nil {
		return err
	}
	return generate(*websiteInfo, *outputDir)
}

func getWebsiteInfo(filePath string) (*wpparser.WebsiteInfo, error) {
	parser := wpparser.NewParser()
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	return parser.Parse(file, strings.Split(*authors, ","))
}

func generate(info wpparser.WebsiteInfo, outputDirPath string) error {
	log.Debug().Msgf("Output: %s", outputDirPath)
	generator := hugogenerator.NewGenerator(outputDirPath, *font, mediacache.New(*mediaCacheDir),
		*downloadMedia, *continueOnMediaDownloadFailure, info)
	return generator.Generate()
}
