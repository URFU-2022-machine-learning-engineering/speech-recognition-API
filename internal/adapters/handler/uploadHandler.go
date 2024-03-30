package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"path/filepath"
	"sr-api/internal/adapters/repository"
	"sr-api/internal/config"
	"sr-api/internal/core/domain"
	"sr-api/internal/core/ports"
	"sr-api/internal/core/ports/telemetry"
)

type UploadHandlerDependencies struct {
	WhisperRepo *repository.WhisperRepository
	MinioRepo   *repository.MinioRepository
}

func NewUploadHandlerDependencies(cfg *config.AppConfig) (*UploadHandlerDependencies, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is nil")
	}

	whisperRepo := repository.NewWhisperRepository(cfg)
	minioRepo, err := repository.NewMinioRepository(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Minio repository: %w", err)
	}

	return &UploadHandlerDependencies{
		WhisperRepo: whisperRepo,
		MinioRepo:   minioRepo,
	}, nil
}

func (dep *UploadHandlerDependencies) UploadHandler(c *gin.Context) {
	_, span := telemetry.StartSpanFromGinContext(c, "UploadHandler")
	defer span.End()

	spanID := telemetry.GetSpanId(span)

	log.Debug().Str("span_id", spanID).Msg("Starting UploadHandler")
	// Extract the file from the request
	file, err := c.FormFile("file")
	if err != nil {
		ports.RespondWithError(c, span, err, http.StatusBadRequest, "Failed to get uploaded file")
		return
	}
	log.Debug().Str("span_id", spanID).Str("file_name", file.Filename).Msg("File extracted from the request")

	openedFile, err := file.Open()
	if err != nil {
		ports.RespondWithError(c, span, err, http.StatusBadRequest, "Failed to open uploaded file")
		return
	}
	defer openedFile.Close()
	log.Debug().Str("span_id", spanID).Msg("Opened file successfully")

	// Check file signature with tracing (adapted to use Gin context)
	if err := domain.CheckFileSignatureWithGinContext(c, openedFile); err != nil {
		ports.RespondWithError(c, span, err, http.StatusBadRequest, "Invalid file signature")
		return
	}
	log.Info().Str("span_id", spanID).Msg("File signature verified")

	// Generate a new filename and attempt to upload
	fileExt := filepath.Ext(file.Filename)
	fileUUID, err := domain.GenerateUIDWithContext(c)
	if err != nil {
		ports.RespondWithError(c, span, err, http.StatusInternalServerError, "Failed to generate UUID for file")
		return
	}
	fileName := fmt.Sprintf("%s%s", fileUUID, fileExt)
	if err := dep.MinioRepo.UploadToMinioWithContext(c, fileName, openedFile, file.Size); err != nil {
		ports.RespondWithError(c, span, err, http.StatusInternalServerError, "Failed to upload file")
		return
	}

	log.Info().Str("span_id", spanID).Str("file_name", fileName).Msg("File uploaded successfully")
	span.AddEvent("File uploaded successfully", trace.WithAttributes(attribute.String("filename", fileName)))
	// Process the file after successful upload
	recognitionResult, err := dep.WhisperRepo.SendToWhisper(c, fileName)
	if err != nil {
		ports.RespondWithError(c, span, err, http.StatusInternalServerError, "Failed to process file")
		return
	}

	c.JSON(http.StatusOK, recognitionResult)
	span.SetStatus(codes.Ok, "File transcribed successfully")
}
