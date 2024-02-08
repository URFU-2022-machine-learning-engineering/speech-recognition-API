package utils

import (
	"context"
	"log"
	"os"

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
		errorMessage := "Critical configuration error: environment variable '' must not be empty."
		log.Println(errorMessage)

		// Record the error in the span
		span.SetStatus(codes.Error, errorMessage)
		span.SetAttributes(attribute.String("environment.variable", key), attribute.String("error.message", errorMessage))

		// Assuming you have a mechanism to flush telemetry data before exiting
		FlushTelemetryData()

		os.Exit(1) // Use os.Exit to terminate the application
	}

	// Correctly use attribute.String to add string attributes to the span
	span.SetAttributes(attribute.String("environment.variable", key), attribute.String("value", value))
	return value
}

// FlushTelemetryData would flush any buffered telemetry data.
// Implement this based on the telemetry system you're using.
func FlushTelemetryData() {
	// Example: Flush data if using a batch span processor or similar
}
