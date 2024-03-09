package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"os"
	"path/filepath"
	"sr-api/utils"
)

func UploadHandler(c *gin.Context) {
	ctx, span := utils.StartSpanFromGinContext(c, "UploadHandler")
	defer span.End()

	log.Info().Msg("Received POST request")

	// Extract the file from the request
	file, err := c.FormFile("file")
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Failed to get uploaded file")
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Failed to open uploaded file")
		return
	}
	defer openedFile.Close()

	// Check file signature with tracing (adapted to use Gin context)
	if err := utils.CheckFileSignatureWithContext(ctx, openedFile); err != nil {
		span.RecordError(err)
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid file signature")
		return
	}

	// Generate a new filename and attempt to upload
	fileExt := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%s%s", utils.GenerateUIDWithContext(ctx), fileExt)
	if err := utils.UploadToMinioWithContext(c, fileName, openedFile, file.Size); err != nil {
		span.RecordError(err)
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to upload file to storage")
		return
	}

	span.AddEvent("File uploaded successfully", trace.WithAttributes(attribute.String("filename", fileName)))

	utils.ProcessFileWithGinContext(c, os.Getenv("MINIO_BUCKET"), fileName)

}
