package repository

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"mime/multipart"
	"sr-api/internal/config"
	"sr-api/internal/core/ports/telemetry"
)

type MinioRepository struct {
	Client *minio.Client
	config *config.AppConfig
}

func NewMinioRepository(cfg *config.AppConfig) (*MinioRepository, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}
	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &MinioRepository{
		Client: minioClient,
		config: cfg,
	}, nil
}

// UploadToMinioWithContext uploads a file to MinIO storage
func (repo *MinioRepository) UploadToMinioWithContext(c *gin.Context, filename string, file multipart.File, size int64) error {
	if repo.config == nil {
		return fmt.Errorf("repository configuration is nil")
	}

	if repo.Client == nil {
		return fmt.Errorf("minio client is nil")
	}

	if file == nil {
		return fmt.Errorf("file is nil")
	}

	ctx, span := telemetry.StartSpanFromGinContext(c, "UploadToMinio")
	bucketName := repo.config.MinioBucket

	if bucketName == "" {
		return fmt.Errorf("bucket name is empty")
	}

	defer span.End()
	spanID := telemetry.GetSpanId(span)

	info, err := repo.Client.PutObject(ctx, bucketName, filename, file, size, minio.PutObjectOptions{})
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
