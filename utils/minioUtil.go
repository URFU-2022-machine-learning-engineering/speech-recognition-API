package utils

import (
	"context"
	"log"
	"mime/multipart"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// UploadToMinioWithContext uploads a file to Minio storage, incorporating context for tracing.
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
		log.Printf("Failed to create Minio client: %v", err)
		return err
	}

	// Specify the bucket name
	bucketName := os.Getenv("MINIO_BUCKET")

	// Perform the upload
	info, err := minioClient.PutObject(ctx, bucketName, filename, file, size, minio.PutObjectOptions{})
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.detail", "Failed to upload file"), attribute.String("minio.bucket", bucketName))
		log.Printf("Failed to upload file to bucket %s: %v", bucketName, err)
		return err
	}

	span.SetAttributes(
		attribute.String("minio.bucket", info.Bucket),
		attribute.String("file.name", filename),
		attribute.Int64("minio.file.size", info.Size),
	)

	log.Printf("Successfully uploaded %s of size %d to bucket %s", filename, size, bucketName)
	return nil
}
