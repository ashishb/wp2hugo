package logger

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	_defaultLogLevel      = zerolog.DebugLevel
	_defaultColoredOutput = false
)

// ConfigureLogging configures ZeroLog's logging config with good defaults
func ConfigureLogging(colorLogOutput bool) {
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logLevel := getLogLevel()
	zerolog.SetGlobalLevel(logLevel)

	if colorLogOutput {
		// Pretty printing is a bit inefficient for production
		output := zerolog.ConsoleWriter{Out: os.Stderr}
		output.FormatTimestamp = func(t any) string {
			ms, err := t.(json.Number).Int64()
			if err != nil {
				panic(err)
			}
			return time.Unix(ms, 0).In(time.Local).Format("03:04:05PM")
		}
		log.Logger = log.Output(output)
		log.Logger = log.With().Caller().Logger()
	}

	zerolog.TimestampFunc = func() time.Time {
		return time.Now().In(time.Local)
	}
	zerolog.CallerMarshalFunc = func(_ uintptr, file string, line int) string {
		// Use just the filename and not the full file path for logging
		fields := strings.Split(file, "/")
		return fields[len(fields)-1] + ":" + strconv.Itoa(line)
	}
}

func getLogLevel() zerolog.Level {
	logLevelStr := strings.TrimSpace(os.Getenv("LOG_LEVEL"))
	if len(logLevelStr) == 0 {
		return _defaultLogLevel
	}
	switch strings.ToUpper(logLevelStr) {
	case "TRACE":
		return zerolog.TraceLevel
	case "DEBUG":
		return zerolog.DebugLevel
	case "INFO":
		return zerolog.InfoLevel
	case "ERROR":
		return zerolog.ErrorLevel
	case "WARN":
		return zerolog.WarnLevel
	case "FATAL":
		return zerolog.FatalLevel
	default:
		panic("Unexpected log level: " + logLevelStr)
	}
}
