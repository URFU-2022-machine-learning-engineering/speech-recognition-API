package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"io"
	"sr-api/helpers"
)

// UploadToMinioWithContext uploads data to MinIO storage, accepting an io.Reader for the data source
func UploadToMinioWithContext(c *gin.Context, filename string, data io.Reader, size int64) error {
	ctx, span := helpers.StartSpanFromGinContext(c, "UploadToMinioWithContext")
	defer span.End()

	// Environment variable retrieval
	minioAccessKey, minioSecretKey, minioEndpoint, bucketName, minioUseSSL := helpers.GetMinIOConfig()

	// MinIO client initialization
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: minioUseSSL == "true",
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Minio client")
		span.RecordError(err)
		return err
	}

	// Upload process
	info, err := minioClient.PutObject(ctx, bucketName, filename, data, size, minio.PutObjectOptions{})
	if err != nil {
		log.Error().Err(err).Str("minio.bucket", bucketName).Msg("Failed to upload data")
		span.RecordError(err)
		return err
	}

	// Success logging and tracing
	span.SetAttributes(
		attribute.String("minio.bucket", info.Bucket),
		attribute.String("file.name", filename),
		attribute.Int64("minio.file.size", info.Size),
	)
	span.SetStatus(codes.Ok, "Data uploaded successfully")
	log.Info().
		Str("file.name", filename).
		Int64("minio.file.size", info.Size).
		Str("minio.bucket", bucketName).
		Msg("Successfully uploaded data to MinIO")

	return nil
}
