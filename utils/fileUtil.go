package utils

import (
	"context"
	"fmt"
	"github.com/h2non/filetype"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"io"
	"mime/multipart"
)

const SignatureLength = 261

// CheckFileSignatureWithContext checks if the given file has a valid file signature.
// It uses context for better tracing, error handling, and integrates zerolog for logging.
func CheckFileSignatureWithContext(ctx context.Context, file multipart.File) error {
	// Start a new span for checking the file signature
	tr := otel.Tracer("file-utils")
	_, span := tr.Start(ctx, "CheckFileSignature")
	defer span.End()

	log.Info().Msg("Starting file signature check")

	// Check file size before reading signature
	size, err := file.Seek(0, io.SeekEnd) // Seek to end of file to get size
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.detail", "Failed to seek file"))
		log.Error().Err(err).Msg("Failed to seek to the end of the file")
		return err
	}
	if size < SignatureLength {
		err := fmt.Errorf("file size too small")
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.detail", "File size too small for signature check"))
		log.Error().Err(err).Msg("File size too small for signature check")
		return err
	}

	// Reset the file pointer to the beginning of the file after checking size
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.detail", "Failed to reset file pointer"))
		log.Error().Err(err).Msg("Failed to reset file pointer to the beginning")
		return err
	}

	log.Info().Msg("File size validation passed, reading file signature")

	// Example file signature checking logic (simplified for demonstration)
	buf := make([]byte, SignatureLength)
	if _, err := file.Read(buf); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.detail", "Failed to read file signature"))
		log.Error().Err(err).Msg("Failed to read file signature")
		return err
	}

	// IMPORTANT: Reset the file pointer to the start again after reading the signature
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.detail", "Failed to reset file pointer after reading signature"))
		log.Error().Err(err).Msg("Failed to reset file pointer to the beginning after reading signature")
		return err
	}

	kind, err := filetype.Match(buf)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.detail", "Error matching file type"))
		log.Error().Err(err).Msg("Error matching file type")
		return err
	}
	if kind == filetype.Unknown {
		err := fmt.Errorf("unknown file type")
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.detail", "Unknown file type"))
		log.Warn().Msg("Unknown file type detected")
		return err
	}

	log.Info().Str("file.type", kind.MIME.Value).Msg("File signature check completed successfully")

	// Annotate the span with the detected file type
	span.SetAttributes(attribute.String("file.type", kind.MIME.Value))

	return nil
}
