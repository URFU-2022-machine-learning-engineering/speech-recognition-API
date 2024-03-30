package ports

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"sr-api/internal/core/ports/telemetry"
)

// RespondWithError sends an error response along with tracing the operation.
func RespondWithError(c *gin.Context, span trace.Span, err error, code int, message string) {
	spanID := telemetry.GetSpanId(span)
	span.SetAttributes(attribute.Int("http.status_code", code), attribute.String("error.message", message))
	span.RecordError(err)

	log.Error().Str("span_id", spanID).Int("http.status_code", code).Msg(message)
	span.SetStatus(codes.Error, message)

	c.JSON(code, gin.H{"error": "something went wrong, please try again later"})
}
