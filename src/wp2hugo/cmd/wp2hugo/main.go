package main

import (
	"flag"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/logger"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/wpparser"
	"github.com/rs/zerolog/log"
	"os"
)

var (
	source = flag.String("source", "", "file path to the source WordPress XML file")
)

func main() {
	flag.Parse()

	// Set log level
	logger.ConfigureLogging()
	if len(*source) == 0 {
		log.Fatal().Msg("Source file is required")
	}
	log.Debug().Msgf("Source: %s", *source)
	err := handle(*source)
	if err != nil {
		log.Fatal().Msgf("Error: %s", err)
	}
}

func handle(filePath string) error {
	parser := wpparser.NewParser()
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	_, err = parser.Parse(file)
	if err != nil {
		return err
	}
	return nil
}
