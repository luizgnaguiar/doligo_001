package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger initializes zerolog with global settings.
// It sets the log level and output format.
func InitLogger(logLevel string) {
	// Set global log level
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		log.Warn().Err(err).Str("provided_level", logLevel).Msg("Invalid log level, defaulting to info.")
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output for human-readable console output during development
	var output io.Writer = os.Stdout
	if zerolog.GlobalLevel() <= zerolog.DebugLevel { // For dev/debug, use pretty console output
		output = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	}

	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	log.Info().Msgf("Logger initialized with level: %s", level.String())
}
