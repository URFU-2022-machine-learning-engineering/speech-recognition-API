package utils

import (
	"context"
	"mime/multipart"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// UploadToMinioWithContext uploads a file to Minio storage, incorporating context for tracing and zerolog for logging.
func UploadToMinioWithContext(ctx context.Context, filename string, file multipart.File, size int64) error {

	_, span := otel.Tracer("file-utils").Start(ctx, "UploadToMinio")
	defer span.End()

	minioAccessKey := GetEnvOrShutdownWithTelemetry(ctx, "MINIO_ACCESS_KEY")
	minioSecretKey := GetEnvOrShutdownWithTelemetry(ctx, "MINIO_SECRET_KEY")
	minioEndpoint := GetEnvOrShutdownWithTelemetry(ctx, "MINIO_ENDPOINT")
	minioUseSSL := os.Getenv("MINIO_USE_SSL") == "true"

	// Initialize Minio client
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: minioUseSSL,
	})
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.detail", "Failed to create Minio client"))
		log.Error().Err(err).Msg("Failed to create Minio client")
		return err
	}

	// Specify the bucket name
	bucketName := os.Getenv("MINIO_BUCKET")

	// Perform the upload
	info, err := minioClient.PutObject(ctx, bucketName, filename, file, size, minio.PutObjectOptions{})
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.detail", "Failed to upload file"), attribute.String("minio.bucket", bucketName))
		log.Error().Err(err).Str("minio.bucket", bucketName).Msg("Failed to upload file")
		return err
	}

	span.SetAttributes(
		attribute.String("minio.bucket", info.Bucket),
		attribute.String("file.name", filename),
		attribute.Int64("minio.file.size", info.Size),
	)

	log.Info().
		Str("file.name", filename).
		Int64("minio.file.size", info.Size).
		Str("minio.bucket", bucketName).
		Msg("Successfully uploaded file")

	return nil
}
