package helpers

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// StartSpanFromGinContext initializes a new tracing span for an HTTP request within a Gin context.
func StartSpanFromGinContext(c *gin.Context, spanName string) (context.Context, trace.Span) {
	tr := otel.Tracer("sr-api")
	ctx, span := tr.Start(c.Request.Context(), spanName, trace.WithAttributes(
		attribute.String("http.method", c.Request.Method),
		attribute.String("http.url", c.Request.URL.String()),
	))
	return ctx, span
}

// StartSpanFromContext starts an OpenTelemetry span from a given context and span name.
// This function is more generic and can be used in both HTTP and WebSocket contexts.
func StartSpanFromContext(ctx context.Context, spanName string, attributes ...attribute.KeyValue) (context.Context, trace.Span) {
	tr := otel.Tracer("sr-api")
	// Start a new span with the provided context and attributes.
	// Additional attributes can be passed to this function as needed.
	newCtx, span := tr.Start(ctx, spanName, trace.WithAttributes(attributes...))
	return newCtx, span
}
