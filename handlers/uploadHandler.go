package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sr-api/utils"

	"github.com/rs/zerolog/log" // Use zerolog for logging

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// Initialize zerolog with context for potential structured logging
	ctx, span := utils.StartSpanFromRequest(r, "UploadHandler")
	defer span.End()

	if r.Method == "POST" {

		// Extract the file from the request, now passing ctx for potential context-aware logging
		file, handler, err := r.FormFile("file")
		if err != nil {
			utils.RespondWithError(ctx, w, http.StatusBadRequest, "Failed to get uploaded file")
			return
		}
		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
				log.Error().Err(err).Msg("Failed to close the file")
			}
		}(file)

		// Check file signature with tracing
		if err := utils.CheckFileSignatureWithContext(ctx, file); err != nil {
			utils.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid file signature")
			return
		}

		// Generate a new filename and attempt to upload, incorporating ctx into the upload function
		fileExt := filepath.Ext(handler.Filename)
		fileName := fmt.Sprintf("%s%s", utils.GenerateRandomStringWithContext(ctx, 16), fileExt)
		if err := utils.UploadToMinioWithContext(ctx, fileName, file, handler.Size); err != nil {
			utils.RespondWithError(ctx, w, http.StatusInternalServerError, "Failed to upload file to storage")
			return
		}

		span.AddEvent("File uploaded successfully", trace.WithAttributes(attribute.String("filename", fileName)))

		// Process the file if necessary, again passing along the context
		result, err := utils.ProcessFileWithContext(ctx, os.Getenv("STORAGE_BUCKET"), fileName)
		if err != nil {
			utils.RespondWithError(ctx, w, http.StatusInternalServerError, "Failed to process file")
			log.Error().Err(err).Msg("Failed to process the file")
			return
		}

		utils.RespondWithSuccess(ctx, w, http.StatusOK, result)
	} else {
		// Log the non-POST request method using zerolog
		log.Warn().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("Remote addr", r.RemoteAddr).
			Msg("Method Not Allowed")

		// Return a 405 Method Not Allowed response for non-POST requests
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, err := fmt.Fprintln(w, "Method Not Allowed")
		if err != nil {
			log.Error().Err(err).Msg("Failed to write Method Not Allowed response")
			return
		}
	}
}
