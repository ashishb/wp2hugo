package main

import (
	"flag"
	"os"
	"path"
	"strings"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugogenerator"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/mediacache"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/rs/zerolog/log"
)

var (
	sourceFile                     = flag.String("source", "", "file path to the source WordPress XML file")
	outputDir                      = flag.String("output", "/tmp", "dir path to write the Hugo-generated data to")
	downloadMedia                  = flag.Bool("download-media", false, "download media files embedded in the WordPress content")
	downloadAll                    = flag.Bool("download-all", false, "download all media from WordPress library, whether used in content or not")
	continueOnMediaDownloadFailure = flag.Bool("continue-on-media-download-error", false, "continue processing even if one or more media downloads fail")
	generateNgnixConfig            = flag.Bool("generate-nginx-config", true, "generate Nginx configuration for the generated Hugo website for redirecting WordPress GUIDs to Hugo URLs")
	authors                        = flag.String("authors", "", "CSV list of author name(s), if provided, only posts by these authors will be processed")
	// This is useful for repeated executions of the tool to avoid downloading the media files again
	// Mostly for development and not for the production use
	mediaCacheDir = flag.String("media-cache-dir", path.Join("/tmp/wp2hugo-cache"), "dir path to cache the downloaded media files")
	// Custom font for Hugo's papermod theme
	font           = flag.String("font", "Lexend", "custom font for the output website")
	colorLogOutput = flag.Bool("color-log-output", true, "enable colored log output, set false to structured JSON log")

	customPostTypes = flag.String("custom-post-types", "", "CSV list of custom post types to import")
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
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0o644)
	if err != nil {
		return nil, err
	}

	defaultCustomPosts := []string{"avada_portfolio", "avada_faq", "product", "product_variation"}
	defaultCustomPosts = append(defaultCustomPosts, strings.Split(*customPostTypes, ",")...)
	
	return parser.Parse(file, strings.Split(*authors, ","), defaultCustomPosts)
}

func generate(info wpparser.WebsiteInfo, outputDirPath string) error {
	log.Debug().Msgf("Output: %s", outputDirPath)
	generator := hugogenerator.NewGenerator(outputDirPath, *font, mediacache.New(*mediaCacheDir),
		*downloadMedia, *downloadAll, *continueOnMediaDownloadFailure, *generateNgnixConfig, info)
	return generator.Generate()
}
