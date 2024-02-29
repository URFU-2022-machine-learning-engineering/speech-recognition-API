package main

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sr-api/handlers" // Ensure this import path is correct and accessible.
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	// Ensure zerolog logs with proper time formatting
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(os.Stderr)
	log.Info().Msg("Starting server...")
	if err := run(); err != nil {
		log.Error().Err(err).Msg("Failed to start server")
	}
}

func run() (err error) {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(ctx)
	log.Info().Msg("OpenTelemetry SDK setup")
	if err != nil {
		log.Error().Err(err).Msg("Setup OTel failed, err")
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()
	srv := &http.Server{
		Addr:         "0.0.0.0:8080",
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHTTPHandler(),
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()
	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		log.Error().Err(err).Msg("HTTP server ListenAndServe")
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		log.Info().Msg("System interruption")
		stop()
	}
	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = srv.Shutdown(context.Background())
	return
}

// Set up HTTP server and routes

func newHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	// Enhanced logging: Log before attempting to register handlers.
	log.Info().Msg("Registering HTTP handlers...")

	// handleFunc is a replacement for mux.HandleFunc
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		// Configure the "http.route" for the HTTP instrumentation.
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
		log.Info().Str("pattern", pattern).Msg("Handler registered")

	}

	// Attempt to register handlers and log any potential errors.
	// Note: Assuming handlers.UploadHandler and handlers.StatusHandler do not return errors.
	// If they can return errors, you should handle those appropriately here.
	handleFunc("/upload", handlers.UploadHandler)
	handleFunc("/status", handlers.StatusHandler)

	log.Info().Msg("All handlers registered successfully.")

	// Add HTTP instrumentation for the whole server.
	handler := otelhttp.NewHandler(mux, "/")
	return handler
}
