package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"sr-api/handlers"
)

func main() {

	setupLogging()
	// Initialize OpenTelemetry
	shutdown, err := setupOTelSDK(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to set up OpenTelemetry")
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Fatal().Err(err).Msg("Failed to shut down OpenTelemetry properly")
		}
	}()

	// Set up your Gin router
	r := gin.New()

	// Use OpenTelemetry middleware to automatically start and end spans for requests
	r.Use(otelgin.Middleware("sr-api"))

	// Set up routes
	r.GET("/status", handlers.StatusHandler)
	r.GET("/ws", handlers.WsHandler)
	r.POST("/upload", handlers.UploadHandler)

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
