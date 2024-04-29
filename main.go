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
	setupLogger()
	shutdown := setupTelemetry()
	defer shutdown(context.Background())

	cfg := config.LoadConfig()
	dep, err := handler.NewUploadHandlerDependencies(cfg)
	checkError(err, "Failed to initialize dependencies")

	r := setupServer()
	r.GET("/status", handler.StatusHandler)
	r.POST("/upload", dep.UploadHandler)

	startServer(r)
}

func setupLogger() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	log.Logger = log.Output(os.Stderr)
}

func setupTelemetry() func(context.Context) error {
	shutdown, err := telemetry.SetupOTelSDK(context.Background())
	checkError(err, "Failed to set up OpenTelemetry")
	return shutdown
}

func setupServer() *gin.Engine {
	r := gin.New()
	r.Use(otelgin.Middleware("sr-api"))
	return r
}

func startServer(r *gin.Engine) {
	err := r.Run(":8080")
	checkError(err, "Failed to start server")
}

func checkError(err error, msg string) {
	if err != nil {
		log.Fatal().Err(err).Msg(msg)
	}
}
