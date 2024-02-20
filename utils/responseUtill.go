package utils

import (
	"context"
	"encoding/json"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"sr-api/handlers/handlers_structure"
)

// RespondWithError sends an error response along with tracing the operation.
func RespondWithError(ctx context.Context, w http.ResponseWriter, code int, message string) {
	// Start a span for the error response
	_, span := otel.Tracer("http-server").Start(ctx, "RespondWithError")
	span.SetAttributes(attribute.Int("http.status_code", code), attribute.String("error.message", message))
	defer span.End()

	log.Printf("Error %d: %s", code, message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	p := handlers_structure.UploadError{Result: message}
	if err := json.NewEncoder(w).Encode(p); err != nil {
		log.Printf("Failed to encode error response JSON: %v", err)
	}
}

// RespondWithSuccess sends a success response and allows for adding tracing information.
func RespondWithSuccess(ctx context.Context, w http.ResponseWriter, code int, payload interface{}) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("RespondWithSuccess", trace.WithAttributes(attribute.Int("http.status_code", code)))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Failed to encode success response JSON: %v", err)
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
		log.Printf("Failed to encode success response JSON: %v", err)
		span.RecordError(err)
	}

}
