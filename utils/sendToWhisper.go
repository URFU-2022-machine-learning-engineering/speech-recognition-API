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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func ProcessFileWithGinContext(c *gin.Context, bucketName, fileName string) {
	ctx, span := helpers.StartSpanFromGinContext(c, "ProcessFileWithGinContext")
	defer span.End()

	whisperEndpoint := helpers.GetEnvOrShutdownWithTelemetry(ctx, "WHISPER_ENDPOINT")
	whisperTranscribe := helpers.GetEnvOrShutdownWithTelemetry(ctx, "WHISPER_TRANSCRIBE")

	u, err := url.Parse(whisperEndpoint)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid Whisper endpoint URL: %v", err)})
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid Whisper endpoint URL")
		return
	}

	u.Path = path.Join(u.Path, whisperTranscribe)
	whisperTranscribeURL := u.String()

	data := handlers_structure.SendData{BucketName: bucketName, FileName: fileName}
	jsonData, err := json.Marshal(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error marshaling request data"})
		span.RecordError(err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", whisperTranscribeURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		span.RecordError(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	// Add other headers like Authorization here

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request to Whisper service"})
		span.RecordError(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "Whisper service returned an error"})
		span.SetStatus(codes.Error, fmt.Sprintf("Whisper service error: %d", resp.StatusCode))
		return
	}

	var recognitionResult handlers_structure.RecognitionSuccess
	if err := json.NewDecoder(resp.Body).Decode(&recognitionResult); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode Whisper service response"})
		span.RecordError(err)
		return
	}

	// If the processing and response are successful, return the decoded JSON response directly.
	c.JSON(http.StatusOK, recognitionResult)
	span.SetAttributes(attribute.String("response.status", "success"))
	span.SetStatus(codes.Ok, "File processed successfully")
}
