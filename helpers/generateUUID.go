package helpers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"

	"go.opentelemetry.io/otel/attribute"
)

// GenerateUIDWithContext generates a unique identifier (UUID),
// optionally tracing the operation.
func GenerateUIDWithContext(c *gin.Context) (string, error) {
	_, span := StartSpanFromGinContext(c, "GenerateUID")
	defer span.End()

	uid, err := uuid.NewUUID()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to generate UUID")
		return "", err
	}
	span.SetStatus(codes.Ok, "UUID generated")
	span.SetAttributes(attribute.String("uuid", uid.String()))

	return uid.String(), nil
}
