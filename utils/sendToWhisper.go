package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"go.opentelemetry.io/otel/codes"
	"io"
	"net/http"
	"net/url"
	"path"
	"sr-api/handlers/handlers_structure"

	"github.com/rs/zerolog/log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// processResponse processes the response body. If it's JSON, it logs and returns the JSON.
// If it's not JSON (assumed HTML), it logs and returns the HTML response.
func processResponse(ctx context.Context, body []byte) (interface{}, error) {
	tracer := otel.Tracer("utils")
	// Start a new span for the processResponse operation
	ctx, span := tracer.Start(ctx, "processResponse")
	defer span.End()

	var response interface{}
	err := json.Unmarshal(body, &response)
	if err == nil {
		// If the unmarshalling is successful, log the JSON response
		log.Info().Msgf("Received JSON response: %v", response)
		// Set span attributes relevant to the operation
		span.SetAttributes(attribute.String("response.type", "json"))
		return response, nil
	}

	// If unmarshalling as JSON fails, treat it as an HTML response
	htmlResponse := string(body)
	// Log the HTML response as an error
	log.Error().Err(err).Msgf("Failed to unmarshal JSON, treating as HTML: %s", htmlResponse)
	// Record the error in the span
	span.RecordError(err)
	span.SetStatus(codes.Error, "Failed to unmarshal JSON")
	span.SetAttributes(
		attribute.String("response.type", "html"),
		attribute.String("error.message", "Failed to unmarshal JSON, treating as HTML"),
	)

	return htmlResponse, nil
}

func ProcessFileWithContext(ctx context.Context, bucketName, fileName string) (interface{}, error) {
	tracer := otel.Tracer("utils")
	ctx, span := tracer.Start(ctx, "ProcessFileWithContext")
	defer span.End()

	whisperEndpoint := GetEnvOrShutdownWithTelemetry(ctx, "WHISPER_ENDPOINT")
	whisperTranscribe := GetEnvOrShutdownWithTelemetry(ctx, "WHISPER_TRANSCRIBE")

	u, err := url.Parse(whisperEndpoint)
	if err != nil {
		span.RecordError(err)
		log.Fatal().Err(err).Msg("Failed to parse WHISPER_ENDPOINT")
	}

	u.Path = path.Join(u.Path, whisperTranscribe)
	whisperTranscribeURL := u.String()

	data := handlers_structure.SendData{BucketName: bucketName, FileName: fileName}
	jsonData, err := json.Marshal(data)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", whisperTranscribeURL, bytes.NewBuffer(jsonData))
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close response body")
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	response, err := processResponse(ctx, body)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("response.error", "Failed to process response"))
		return nil, err
	}

	return response, nil
}
