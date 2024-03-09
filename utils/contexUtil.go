package utils

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// StartSpanFromGinContext initializes a new tracing span for an HTTP request within a Gin context.
func StartSpanFromGinContext(c *gin.Context, spanName string) (context.Context, trace.Span) {
	tr := otel.Tracer("sr-api") // Change "your-service-name" to the actual name of your service.
	ctx, span := tr.Start(c.Request.Context(), spanName, trace.WithAttributes(
		attribute.String("http.method", c.Request.Method),
		attribute.String("http.url", c.Request.URL.String()),
	))
	return ctx, span
}
