package helpers

import (
	"context"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// GenerateUIDWithContext generates a unique identifier (UUID),
// optionally tracing the operation.
func GenerateUIDWithContext(ctx context.Context) string {

	_, span := otel.Tracer("utils").Start(ctx, "GenerateUID")
	defer span.End()

	uid, err := uuid.NewUUID()
	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return "" // or handle the error according to your application's requirements
	}
	return uid.String()
}
