package utils

import (
	"context"
	"encoding/json"
	"net/http"
	"sr-api/handlers/handlers_structure"

	"github.com/rs/zerolog/log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// RespondWithError sends an error response along with tracing the operation.
func RespondWithError(ctx context.Context, w http.ResponseWriter, code int, message string) {
	// Start a span for the error response
	_, span := otel.Tracer("http-server").Start(ctx, "RespondWithError")
	span.SetAttributes(attribute.Int("http.status_code", code), attribute.String("error.message", message))
	defer span.End()

	// Use zerolog for logging the error
	log.Error().Int("code", code).Str("message", message).Msg("Error response")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	p := handlers_structure.UploadError{Result: message}
	if err := json.NewEncoder(w).Encode(p); err != nil {
		// Log the failure to encode the error response
		log.Error().Err(err).Msg("Failed to encode error response JSON")
	}
}

// RespondWithSuccess sends a success response and allows for adding tracing information.
func RespondWithSuccess(ctx context.Context, w http.ResponseWriter, code int, payload interface{}) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("RespondWithSuccess", trace.WithAttributes(attribute.Int("http.status_code", code)))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// Log the failure to encode the success response
		log.Error().Err(err).Msg("Failed to encode success response JSON")
		span.RecordError(err)
	}
}

func RespondWithInfo(ctx context.Context, w http.ResponseWriter, code int) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("RespondInfo", trace.WithAttributes(attribute.Int("http.status_code", code)))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	p := handlers_structure.StatusSuccess{Status: "Server is running"}
	if err := json.NewEncoder(w).Encode(p); err != nil {
		// Log the failure to encode the info response
		log.Error().Err(err).Msg("Failed to encode info response JSON")
		span.RecordError(err)
	}
}
