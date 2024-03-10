package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sr-api/handlers/handlers_structure"
	"sr-api/helpers"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func ProcessFileWithGinContext(c *gin.Context, bucketName, fileName string) {
	ctx, span := helpers.StartSpanFromGinContext(c, "ProcessFileWithGinContext")
	defer span.End()

	log.Debug().Str("bucket", bucketName).Str("file", fileName).Msg("Initiating file processing")

	whisperEndpoint := helpers.GetEnvOrShutdownWithTelemetry(c, "WHISPER_ENDPOINT")
	whisperTranscribe := helpers.GetEnvOrShutdownWithTelemetry(c, "WHISPER_TRANSCRIBE")

	u, err := url.Parse(whisperEndpoint)
	if err != nil {
		log.Error().Err(err).Msg("Invalid Whisper endpoint URL")
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid Whisper endpoint URL: %v", err)})
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid Whisper endpoint URL")
		return
	}

	u.Path = path.Join(u.Path, whisperTranscribe)
	whisperTranscribeURL := u.String()

	log.Debug().Str("whisperTranscribeURL", whisperTranscribeURL).Msg("Whisper transcribe URL constructed")

	data := handlers_structure.SendData{BucketName: bucketName, FileName: fileName}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Msg("Error marshaling request data")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error marshaling request data"})
		span.RecordError(err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", whisperTranscribeURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create request")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive")
	// Add other headers like Authorization here if needed

	log.Debug().Msg("Sending request to Whisper service")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send request to Whisper service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request to Whisper service"})
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to send request to Whisper service")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().Int("status_code", resp.StatusCode).Msg("Whisper service returned an error")
		c.JSON(resp.StatusCode, gin.H{"error": "Whisper service returned an error"})
		span.SetStatus(codes.Error, fmt.Sprintf("Whisper service error: %d", resp.StatusCode))
		return
	}

	var recognitionResult handlers_structure.RecognitionSuccess
	if err := json.NewDecoder(resp.Body).Decode(&recognitionResult); err != nil {
		log.Error().Err(err).Msg("Failed to decode Whisper service response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode Whisper service response"})
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to decode Whisper service response")
		return
	}

	log.Info().Str("file", fileName).Msg("File processing completed successfully")
	c.JSON(http.StatusOK, recognitionResult)
	span.SetAttributes(attribute.String("response.status", "success"))
	span.SetStatus(codes.Ok, "File processed successfully")
}
