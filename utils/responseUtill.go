package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// RespondWithError sends an error response along with tracing the operation.
func RespondWithError(c *gin.Context, code int, message string) {
	_, span := otel.Tracer("http-server").Start(c.Request.Context(), "RespondWithError")
	span.SetAttributes(attribute.Int("http.status_code", code), attribute.String("error.message", message))
	defer span.End()

	log.Error().Int("code", code).Str("message", message).Msg("Error response")

	c.JSON(code, gin.H{"error": message})
}
