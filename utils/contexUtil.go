package utils

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// StartSpanFromRequest initializes a new tracing span for an HTTP request.
// It returns the newly created span along with the context that has been updated to include the span.
func StartSpanFromRequest(r *http.Request, spanName string) (context.Context, trace.Span) {
	tr := otel.Tracer("whisper-speech-recognition-api")
	ctx, span := tr.Start(r.Context(), spanName, trace.WithAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.url", r.URL.String()),
	))
	return ctx, span
}
