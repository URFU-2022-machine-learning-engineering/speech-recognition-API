package helpers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/h2non/filetype"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"io"
)

const SignatureLength = 261

// CheckFileSignatureWithGinContext checks if the given reader (file or bytes) has a valid file signature
// within a Gin application context, utilizing tracing.
func CheckFileSignatureWithGinContext(c *gin.Context, reader io.Reader) (string, error) {
	_, span := StartSpanFromGinContext(c, "CheckFileSignatureWithGinContext")
	defer span.End()

	log.Debug().Msg("Initiating file signature verification process")

	log.Debug().Msg("Reading content signature")
	buf := make([]byte, SignatureLength)
	if _, err := io.ReadFull(reader, buf); err != nil {
		logAndSpanError(span, err, "Failed to read content signature")
		return "", err
	}

	// Attempt to reset the reader if it supports seeking
	if seeker, ok := reader.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			logAndSpanError(span, err, "Failed to reset reader to start after signature check")
			return "", err
		}
	} else {
		log.Debug().Msg("Reader does not support seeking, unable to reset after signature check")
		return "", fmt.Errorf("reader does not support seeking, unable to reset after signature check")
	}

	kind, err := filetype.Match(buf)
	if err != nil {
		logAndSpanError(span, err, "Error matching content type")
		return "", err
	}
	if kind == filetype.Unknown {
		err := fmt.Errorf("unknown content type")
		logAndSpanError(span, err, "Content type is unknown")
		return "", err
	}
	contentType := kind.MIME.Value
	log.Info().Str("content_type", kind.MIME.Value).Msg("Content signature verification successful")
	span.SetAttributes(attribute.String("content.type", kind.MIME.Value))
	span.SetStatus(codes.Ok, "Content signature verified successfully")

	return contentType, nil
}
