package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func setupLogging() {
	// Set global log level to debug
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	log.Logger = log.Output(os.Stderr)
	// Ensure zerolog logs with proper time formatting
}
