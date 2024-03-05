package utils

import (
	"context"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/codes"
	"net/http"
	"sr-api/handlers/handlers_structure"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// RespondWithError sends an error response along with tracing the operation.
func RespondWithError(ctx context.Context, w http.ResponseWriter, code int, message string) {
	_, span := otel.Tracer("http-server").Start(ctx, "RespondWithError")
	span.SetAttributes(attribute.Int("http.status_code", code), attribute.String("error.message", message))
	defer span.End()

	log.Error().Int("code", code).Str("message", message).Msg("Error response")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	p := handlers_structure.UploadError{Result: message}
	if err := json.NewEncoder(w).Encode(p); err != nil {
		log.Error().Err(err).Msg("Failed to encode error response JSON")
		span.RecordError(err) // Record the error in the span
	}
}

// RespondWithSuccess sends a success response and allows for adding tracing information.
func RespondWithSuccess(ctx context.Context, w http.ResponseWriter, code int, payload []byte) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("RespondWithSuccess", trace.WithAttributes(attribute.Int("http.status_code", code)))
	defer span.End()

	//w.Header().Set("Content-Type", "application/json")
	//w.Header().Set("Access-Control-Allow-Origin", "*.dzailz.su")
	//w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
	span.SetStatus(codes.Ok, "Successfully parsed response")
	//w.WriteHeader(code)
	log.Debug().Msgf("Written header with code: %d", code)
	_, err := w.Write(payload)
	log.Debug().Msgf("Writren payload: %s", string(payload))
	if err != nil {
		log.Error().Err(err).Msg("Failed to write response payload")
		span.RecordError(err)
	}
	w.(http.Flusher).Flush()
	span.SetStatus(codes.Ok, "Response sent successfully")
}

func RespondWithInfo(ctx context.Context, w http.ResponseWriter, code int) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("RespondInfo", trace.WithAttributes(attribute.Int("http.status_code", code)))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	p := handlers_structure.StatusSuccess{Status: "Server is running"}
	if err := json.NewEncoder(w).Encode(p); err != nil {
		log.Error().Err(err).Msg("Failed to encode info response JSON")
		span.RecordError(err)
	}
}
