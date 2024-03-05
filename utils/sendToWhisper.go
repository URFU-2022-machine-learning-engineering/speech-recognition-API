package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

// processResponse processes the response body. It attempts to unmarshal the body into JSON.
// If the unmarshalling is unsuccessful, it logs the response as a string for debugging.
func processResponse(ctx context.Context, body []byte) ([]byte, error) {
	tracer := otel.Tracer("utils")
	_, span := tracer.Start(ctx, "processResponse")
	defer span.End()

	var response json.RawMessage // Use RawMessage for delayed unmarshalling
	err := json.Unmarshal(body, &response)
	if err != nil {
		// If the body cannot be unmarshalled into JSON, log and return an error
		log.Error().Err(err).Str("response", string(body)).Msg("Failed to unmarshal JSON response")
		span.SetAttributes(attribute.String("response.type", "invalid-json"))
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to unmarshal JSON response")
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	log.Info().Msg("Received JSON response")
	span.SetAttributes(attribute.String("response.type", "json"))
	span.SetStatus(codes.Ok, "Successfully processed response")
	return body, nil // Return the original body for further processing or forwarding
}

func ProcessFileWithContext(ctx context.Context, bucketName, fileName string) ([]byte, error) {
	tracer := otel.Tracer("utils")
	ctx, span := tracer.Start(ctx, "ProcessFileWithContext")
	defer span.End()
	log.Info().Msgf("Processing file: %s", fileName)

	whisperEndpoint := GetEnvOrShutdownWithTelemetry(ctx, "WHISPER_ENDPOINT")
	whisperTranscribe := GetEnvOrShutdownWithTelemetry(ctx, "WHISPER_TRANSCRIBE")

	u, err := url.Parse(whisperEndpoint)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse whisper endpoint URL")
		span.RecordError(err)
		return nil, err
	}

	u.Path = path.Join(u.Path, whisperTranscribe)
	whisperTranscribeURL := u.String()
	log.Debug().Msgf("Whisper transcribe URL: %s", whisperTranscribeURL)
	data := handlers_structure.SendData{BucketName: bucketName, FileName: fileName}
	log.Debug().Msgf("Data: %v", data)
	jsonData, err := json.Marshal(data)
	log.Debug().Msgf("JSON data: %s", jsonData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal JSON data")
		span.RecordError(err)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", whisperTranscribeURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create new request")
		span.RecordError(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Authorization", "Bearer "+GetEnvOrShutdownWithTelemetry(ctx, "WHISPER_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send request")
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
		log.Error().Err(err).Msg("Failed to read response body")
		span.RecordError(err)
		return nil, err
	}

	response, err := processResponse(ctx, body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to process response")
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to process response")
		return nil, err
	}

	span.SetStatus(codes.Ok, "File processed successfully")
	return response, nil
}
