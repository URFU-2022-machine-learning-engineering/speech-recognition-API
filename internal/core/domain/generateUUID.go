package domain

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/codes"
	"sr-api/internal/core/ports/telemetry"

	"go.opentelemetry.io/otel/attribute"
)

// GenerateUIDWithContext generates a unique identifier (UUID),
// optionally tracing the operation.
func GenerateUIDWithContext(c *gin.Context) (string, error) {
	_, span := telemetry.StartSpanFromGinContext(c, "GenerateUID")
	spanID := telemetry.GetSpanId(span)
	defer span.End()
	log.Debug().Str("span_id", spanID).Msg("Generating UUID")

	uid, err := uuid.NewUUID()
	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return "", err
	}
	span.SetStatus(codes.Ok, "UUID generated")
	span.SetAttributes(attribute.String("uuid", uid.String()))
	log.Info().Str("span_id", spanID).Str("uuid", uid.String()).Msg("UUID generated")
	return uid.String(), nil
}
