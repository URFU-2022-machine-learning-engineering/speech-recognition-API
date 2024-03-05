package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sr-api/utils"

	"github.com/rs/zerolog/log"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// Initialize zerolog with context for potential structured logging
	ctx, span := utils.StartSpanFromRequest(r, "UploadHandler")
	defer span.End()

	// Ensure that the method is POST, simplify method check and immediate return on failure
	if r.Method != http.MethodPost {
		span.AddEvent("Method Not Allowed", trace.WithAttributes(attribute.String("method", r.Method)))
		log.Warn().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("Method Not Allowed")

		utils.RespondWithError(ctx, w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	log.Info().Msg("Received POST request")

	// Extract the file from the request, now passing ctx for potential context-aware logging
	file, handler, err := r.FormFile("file")
	if err != nil {
		utils.RespondWithError(ctx, w, http.StatusBadRequest, "Failed to get uploaded file")
		return
	}
	defer file.Close()

	// Check file signature with tracing
	if err := utils.CheckFileSignatureWithContext(ctx, file); err != nil {
		span.RecordError(err)
		utils.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid file signature")
		return
	}

	// Generate a new filename and attempt to upload, incorporating ctx into the upload function
	fileExt := filepath.Ext(handler.Filename)
	fileName := fmt.Sprintf("%s%s", utils.GenerateUIDWithContext(ctx), fileExt)
	if err := utils.UploadToMinioWithContext(ctx, fileName, file, handler.Size); err != nil {
		span.RecordError(err)
		utils.RespondWithError(ctx, w, http.StatusInternalServerError, "Failed to upload file to storage")
		return
	}

	span.AddEvent("File uploaded successfully", trace.WithAttributes(attribute.String("filename", fileName)))

	result, err := utils.ProcessFileWithContext(ctx, os.Getenv("MINIO_BUCKET"), fileName)
	if err != nil {
		utils.RespondWithError(ctx, w, http.StatusInternalServerError, "Failed to process file")
		span.RecordError(err)
		log.Error().Err(err).Msg("Failed to process the file")
		return
	}

	log.Debug().Msgf("Processing Result: %+v", result) // Changed log message for clarity

	// Serialize result to JSON before sending
	//jsonResponse, err := json.Marshal(result)
	if err != nil {
		utils.RespondWithError(ctx, w, http.StatusInternalServerError, "Failed to serialize response")
		log.Error().Err(err).Msg("Failed to serialize the processing result")
		return
	}

	utils.RespondWithSuccess(ctx, w, http.StatusOK, result)
	return
}
