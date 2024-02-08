package utils

import (
	"context"
	"fmt"
	"github.com/h2non/filetype"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"mime/multipart"
)

const SignatureLength = 261

// CheckFileSignatureWithContext checks if the given file has a valid file signature.
// It is updated to use context for better tracing and error handling.
func CheckFileSignatureWithContext(ctx context.Context, file multipart.File) error {
	// Start a new span for checking the file signature
	_, span := otel.Tracer("file-utils").Start(ctx, "CheckFileSignature")
	defer span.End()

	// Check file size before reading signature
	size, err := file.Seek(0, 2) // Seek to end of file to get size
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.detail", "Failed to seek file"))
		return err
	}
	if size < SignatureLength {
		err := fmt.Errorf("file size too small")
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.detail", "File size too small for signature check"))
		return err
	}

	// Ensure to reset the file pointer to the beginning of the file after seeking
	if _, err := file.Seek(0, 0); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.detail", "Failed to reset file pointer"))
		return err
	}

	// Example file signature checking logic (simplified for demonstration)
	buf := make([]byte, SignatureLength)
	if _, err := file.Read(buf); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.detail", "Failed to read file signature"))
		return err
	}

	kind, _ := filetype.Match(buf)
	if kind == filetype.Unknown {
		err := fmt.Errorf("unknown file type")
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.detail", "Unknown file type"))
		return err
	}

	// Add more logic as needed, and don't forget to annotate the span with useful attributes
	span.SetAttributes(attribute.String("file.type", kind.MIME.Value))

	return nil
}
