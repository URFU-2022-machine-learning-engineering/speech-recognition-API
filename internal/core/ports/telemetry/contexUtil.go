package telemetry

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// StartSpanFromGinContext initializes a new tracing span for an HTTP request within a Gin context.
func StartSpanFromGinContext(c *gin.Context, spanName string) (context.Context, trace.Span) {
	log.Debug().Msgf("Starting span '%s' from Gin context", spanName)
	tr := otel.Tracer("sr-api")
	ctx, span := tr.Start(c.Request.Context(), spanName, trace.WithAttributes(
		attribute.String("http.method", c.Request.Method),
		attribute.String("http.url", c.Request.URL.String()),
	))
	return ctx, span
}

func GetSpanId(span trace.Span) string {
	log.Debug().Msg("Getting span ID")
	spanContext := span.SpanContext()
	spanID := spanContext.SpanID().String()
	return spanID
}
