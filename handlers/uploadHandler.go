package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"os"
	"path/filepath"
	"sr-api/helpers"
	"sr-api/utils"
)

func UploadHandler(c *gin.Context) {
	_, span := helpers.StartSpanFromGinContext(c, "UploadHandler")
	defer span.End()

	log.Debug().Msg("Starting UploadHandler")

	// Log the start of the request handling
	log.Info().Msg("Received POST request for file upload")

	// Extract the file from the request
	file, err := c.FormFile("file")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get uploaded file from request")
		utils.RespondWithError(c, http.StatusBadRequest, "Failed to get uploaded file")
		return
	}
	log.Debug().Str("file_name", file.Filename).Msg("File extracted from the request")

	openedFile, err := file.Open()
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Failed to open uploaded file")
		return
	}
	defer openedFile.Close()
	log.Debug().Msg("Opened file successfully")

	// Check file signature with tracing (adapted to use Gin context)
	if _, err := helpers.CheckFileSignatureWithGinContext(c, openedFile); err != nil {
		span.RecordError(err)
		log.Error().Err(err).Msg("Invalid file signature")
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid file signature")
		return
	}
	log.Info().Msg("File signature verified")

	// Generate a new filename and attempt to upload
	fileExt := filepath.Ext(file.Filename)
	fileUUID, err := helpers.GenerateUIDWithContext(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate UUID for file")
		utils.RespondWithError(c, http.StatusInternalServerError, "Server error")
		return
	}
	fileName := fmt.Sprintf("%s%s", fileUUID, fileExt)
	if err := utils.UploadToMinioWithContext(c, fileName, openedFile, file.Size); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to upload file to storage")
		log.Error().Err(err).Str("file_name", fileName).Msg("Failed to upload file to storage")
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to upload file to storage")
		return
	}

	log.Info().Str("file_name", fileName).Msg("File uploaded successfully")
	span.AddEvent("File uploaded successfully", trace.WithAttributes(attribute.String("filename", fileName)))

	// Process the file after successful upload
	utils.ProcessFileWithGinContext(c, os.Getenv("MINIO_BUCKET"), fileName)
	log.Debug().Str("bucket", os.Getenv("MINIO_BUCKET")).Str("file_name", fileName).Msg("Initiated file processing")
	span.SetStatus(codes.Ok, "Successfully uploaded the file")
}
