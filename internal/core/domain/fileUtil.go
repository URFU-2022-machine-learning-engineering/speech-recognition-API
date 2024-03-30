package domain

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/h2non/filetype"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"io"
	"mime/multipart"
	"sr-api/internal/core/ports/telemetry"
)

const SignatureLength = 261

// CheckFileSignatureWithGinContext checks if the given file has a valid file signature
// within a Gin application context, utilizing tracing.
func CheckFileSignatureWithGinContext(c *gin.Context, file multipart.File) error {
	_, span := telemetry.StartSpanFromGinContext(c, "CheckFileSignatureWithGinContext")
	defer span.End()

	// Retrieve the Span ID
	spanID := telemetry.GetSpanId(span)

	log.Debug().Str("span_id", spanID).Msg("Initiating file signature verification process")

	// File size check...
	size, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		logAndSpanError(span, spanID, err, "Failed to seek file to end")
		return err
	}
	if size < SignatureLength {
		err := fmt.Errorf("file size is too small")
		logAndSpanError(span, spanID, err, "File size too small for signature check")
		return err
	}
	log.Debug().Str("span_id", spanID).Msg("File size validation passed")

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		logAndSpanError(span, spanID, err, "Failed to reset file pointer to start")
		return err
	}

	log.Debug().Str("span_id", spanID).Msg("Reading file signature")
	buf := make([]byte, SignatureLength)
	if _, err := file.Read(buf); err != nil {
		logAndSpanError(span, spanID, err, "Failed to read file signature")
		return err
	}

	// Reset the file pointer to the start after reading the signature
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		logAndSpanError(span, spanID, err, "Failed to reset file pointer after signature read")
		return err
	}

	kind, err := filetype.Match(buf)
	if err != nil {
		logAndSpanError(span, spanID, err, "Error matching file type")
		return err
	}
	if kind == filetype.Unknown {
		err := fmt.Errorf("unknown file type")
		logAndSpanError(span, spanID, err, "File type is unknown")
		return err
	}

	log.Info().Str("span_id", spanID).Str("file_type", kind.MIME.Value).Msg("File signature verification successful")
	span.SetAttributes(attribute.String("file.type", kind.MIME.Value))
	span.SetStatus(codes.Ok, "File signature verified successfully")

	return nil
}

func logAndSpanError(span trace.Span, spanID string, err error, message string) {
	log.Error().Str("span_id", spanID).Err(err).Msg(message)
	span.RecordError(err)
	span.SetAttributes(attribute.String("error.detail", message))
}
