package helpers

import (
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func logAndSpanError(span trace.Span, err error, message string) {
	log.Error().Err(err).Msg(message)
	span.RecordError(err)
	span.SetStatus(codes.Error, message)
}
