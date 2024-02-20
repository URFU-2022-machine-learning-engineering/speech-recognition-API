package utils

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"math/rand"
)

// GenerateRandomStringWithContext generates a random string of a specified length,
// optionally tracing the operation.
func GenerateRandomStringWithContext(ctx context.Context, length int) string {
	// Start a new span for this operation, if tracing is desired.
	// This is optional and can be skipped based on performance considerations.
	_, span := otel.Tracer("utils").Start(ctx, "GenerateRandomString")
	span.SetAttributes(attribute.Int("length", length))
	defer span.End()

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
