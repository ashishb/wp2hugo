package main

import (
	"flag"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugogenerator"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/rs/zerolog/log"
	"os"
)

var (
	sourceFile     = flag.String("source", "", "file path to the source WordPress XML file")
	mediaSourceURL = flag.String("media-source", "", "URL to WordPress website to download media files")
	outputDir      = flag.String("output", "/tmp", "dir path to the write the Hugo generated data to")
)

func main() {
	flag.Parse()

	// Set log level
	logger.ConfigureLogging()
	if len(*sourceFile) == 0 {
		log.Fatal().Msg("Source file is required")
	}
	if len(*outputDir) == 0 {
		log.Fatal().Msg("Output directory is required")
	}
	if len(*mediaSourceURL) == 0 {
		log.Warn().Msg("Media source is not provided. Media files will not be downloaded")
	}
	err := handle(*sourceFile, *mediaSourceURL)
	if err != nil {
		log.Fatal().Msgf("Error: %s", err)
	}
}

func handle(filePath string, mediaSourceURL string) error {
	log.Debug().Msgf("Source: %s", filePath)
	websiteInfo, err := getWebsiteInfo(filePath)
	if err != nil {
		return err
	}
	return generate(*websiteInfo, mediaSourceURL, *outputDir)
}

func getWebsiteInfo(filePath string) (*wpparser.WebsiteInfo, error) {
	parser := wpparser.NewParser()
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	return parser.Parse(file)
}

func generate(info wpparser.WebsiteInfo, mediaSourceURL string, outputDirPath string) error {
	log.Debug().Msgf("Output: %s", outputDirPath)
	generator := hugogenerator.NewGenerator()
	return generator.Generate(info, mediaSourceURL, outputDirPath)
}
