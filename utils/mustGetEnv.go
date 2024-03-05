package utils

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// GetEnvOrShutdownWithTelemetry retrieves the value of the environment variable named by the key.
// It logs a critical error, records a trace span, and terminates the application if the variable is empty or not set.
func GetEnvOrShutdownWithTelemetry(ctx context.Context, key string) string {
	tracer := otel.Tracer("app_or_package_name")
	_, span := tracer.Start(ctx, "GetEnvOrShutdownWithTelemetry")
	defer span.End()

	value := os.Getenv(key)
	if value == "" {
		errorMessage := "Critical configuration error: environment variable '" + key + "' must not be empty."
		log.Error().Str("environment.variable", key).Msg(errorMessage)

		// Record the error in the span
		span.SetStatus(codes.Error, errorMessage)
		span.SetAttributes(attribute.String("environment.variable", key), attribute.String("error.message", errorMessage))

		os.Exit(1) // Use os.Exit to terminate the application
	}

	// Correctly use attribute.String to add string attributes to the span
	span.SetStatus(codes.Ok, "Environment variable found")
	return value
}
