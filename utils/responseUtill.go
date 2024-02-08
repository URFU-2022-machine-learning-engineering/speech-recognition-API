package utils

import (
	"context"
	"dzailz.ru/api/handlers/handlers_structure"
	"encoding/json"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
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
	// Optionally, add tracing information related to the success response
	span := trace.SpanFromContext(ctx)
	span.AddEvent("RespondWithSuccess", trace.WithAttributes(attribute.Int("http.status_code", code)))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Failed to encode success response JSON: %v", err)
		// Consider logging the error in the span as well
		span.RecordError(err)
	}
}

func RespondWithInfo(ctx context.Context, w http.ResponseWriter, code int) {
	// Optionally, add tracing information related to the success response
	span := trace.SpanFromContext(ctx)
	span.AddEvent("RespondInfo", trace.WithAttributes(attribute.Int("http.status_code", code)))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	p := handlers_structure.StatusSuccess{Status: "Server is running"}
	if err := json.NewEncoder(w).Encode(p); err != nil {
		log.Printf("Failed to encode success response JSON: %v", err)
		// Consider logging the error in the span as well
		span.RecordError(err)
	}

}
