package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"mime/multipart"
	"net/http"
	"os"
	"sr-api/helpers"
)

// GetEnv retrieves an environment variable or sends an error response if not set.
func GetEnv(c *gin.Context, key string) (string, bool) {
	_, span := helpers.StartSpanFromGinContext(c, "GetEnv")
	spanID := helpers.GetSpanId(span)
	value := os.Getenv(key)
	if value == "" {
		log.Fatal().Str("span_id", spanID).Str("environment.variable", key).Msg(key + " environment variable is not set")
		span.SetAttributes(attribute.String("fatal", key+" environment variable is not set"))
		span.SetStatus(codes.Error, key+" environment variable is not set")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong, please try again later"})
		return "", false
	}
	return value, true
}

// UploadToMinioWithContext uploads a file to MinIO storage
func UploadToMinioWithContext(c *gin.Context, filename string, file multipart.File, size int64) error {
	ctx, span := helpers.StartSpanFromGinContext(c, "UploadToMinio")
	defer span.End()

	// Use the helper function to get the Span ID
	spanID := helpers.GetSpanId(span)

	minioAccessKey, ok := GetEnv(c, "MINIO_ACCESS_KEY")
	if !ok {
		return nil // Since GetEnv sends its own JSON response, simply return here.
	}
	minioSecretKey, ok := GetEnv(c, "MINIO_SECRET_KEY")
	if !ok {
		return nil // Early return after GetEnv has handled the response.
	}
	minioEndpoint, ok := GetEnv(c, "MINIO_ENDPOINT")
	if !ok {
		return nil
	}
	minioUseSSL := os.Getenv("MINIO_USE_SSL") == "true"
	bucketName, ok := GetEnv(c, "MINIO_BUCKET")
	if !ok {
		return nil
	}

	// Initialize Minio client
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: minioUseSSL,
	})
	if err != nil {
		log.Error().Str("span_id", spanID).Err(err).Msg("Failed to create Minio client")
		span.RecordError(err)
		return err
	}

	// Perform the upload
	info, err := minioClient.PutObject(ctx, bucketName, filename, file, size, minio.PutObjectOptions{})
	if err != nil {
		log.Error().Str("span_id", spanID).Err(err).Str("minio.bucket", bucketName).Msg("Failed to upload file")
		span.RecordError(err)
		return err
	}

	span.SetAttributes(
		attribute.String("minio.bucket", info.Bucket),
		attribute.String("file.name", filename),
		attribute.Int64("minio.file.size", info.Size),
	)
	span.SetStatus(codes.Ok, "File uploaded successfully")
	log.Info().
		Str("span_id", spanID).
		Str("file.name", filename).
		Int64("minio.file.size", info.Size).
		Str("minio.bucket", bucketName).
		Msg("Successfully uploaded file to MinIO")

	return nil
}
