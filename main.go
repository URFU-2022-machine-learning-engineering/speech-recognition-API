package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"os"
	"sr-api/internal/adapters/handler"
	"sr-api/internal/config"
	"sr-api/internal/core/ports/telemetry"
)

func main() {

	// Set global log level to debug
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	log.Logger = log.Output(os.Stderr)
	// Initialize OpenTelemetry
	shutdown, err := telemetry.SetupOTelSDK(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to set up OpenTelemetry")
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Fatal().Err(err).Msg("Failed to shut down OpenTelemetry properly")
		}
	}()
	log.Debug().Msg("Loading configuration...")
	cfg := config.LoadConfig()
	if cfg == nil {
		log.Fatal().Msg("Configuration is nil after loading")
	}

	dep, err := handler.NewUploadHandlerDependencies(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize dependencies")
	}

	log.Debug().Msg("Initializing server...")
	r := gin.New()
	log.Debug().Msg("Setting up OpenTelemetry middleware")
	r.Use(otelgin.Middleware("sr-api"))

	log.Debug().Msg("Setting up routes")
	r.GET("/status", handler.StatusHandler)
	r.POST("/upload", dep.UploadHandler)

	log.Debug().Msg("Starting server...")
	if err := r.Run(":8080"); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
