package repository

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"net/http"
	"net/url"
	"path"
	"sr-api/internal/adapters/handler/handlerStructure"
	"sr-api/internal/config"
	http2 "sr-api/internal/core/ports"
	"sr-api/internal/core/ports/telemetry"
)

type WhisperRepository struct {
	config *config.AppConfig
}

func NewWhisperRepository(cfg *config.AppConfig) *WhisperRepository {
	return &WhisperRepository{
		config: cfg,
	}
}

func (repo *WhisperRepository) SendToWhisper(c *gin.Context, fileName string) (handlerStructure.RecognitionSuccess, error) {
	ctx, span := telemetry.StartSpanFromGinContext(c, "SendToWhisper")
	defer span.End()
	// Now using config from repo
	whisperEndpoint := repo.config.WhisperEndpoint
	whisperTranscribe := repo.config.WhisperTranscribe
	minioBucketName := repo.config.MinioBucket

	spanID := telemetry.GetSpanId(span)

	log.Debug().Str("span_id", spanID).Str("bucket", minioBucketName).Str("file", fileName).Msg("Initiating file processing")

	u, err := url.Parse(whisperEndpoint)
	if err != nil {
		log.Error().Str("span_id", spanID).Err(err).Msg("Invalid Whisper endpoint URL")
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid Whisper endpoint URL")
		return handlerStructure.RecognitionSuccess{}, err
	}

	u.Path = path.Join(u.Path, whisperTranscribe)
	whisperTranscribeURL := u.String()

	log.Debug().Str("span_id", spanID).Str("whisperTranscribeURL", whisperTranscribeURL).Msg("Whisper transcribe URL constructed")

	data := handlerStructure.SendData{BucketName: minioBucketName, FileName: fileName}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error().Str("span_id", spanID).Err(err).Msg("Error marshaling request data")
		span.RecordError(err)
		span.SetStatus(codes.Error, "Error marshaling request data")
		return handlerStructure.RecognitionSuccess{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", whisperTranscribeURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error().Str("span_id", spanID).Err(err).Msg("Failed to create request")
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create request")
		return handlerStructure.RecognitionSuccess{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive")

	log.Debug().Str("span_id", spanID).Msg("Sending request to Whisper service")

	resp := http2.HttpClient(c, req)
	defer resp.Body.Close()

	var recognitionResult handlerStructure.RecognitionSuccess
	if err := json.NewDecoder(resp.Body).Decode(&recognitionResult); err != nil {
		log.Error().Str("span_id", spanID).Err(err).Msg("Failed to decode Whisper service response")
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to decode Whisper service response")
		return handlerStructure.RecognitionSuccess{}, err
	}

	log.Info().Str("span_id", spanID).Str("file", fileName).Msg("File processing completed successfully")
	span.SetAttributes(attribute.String("response.status", "success"))
	span.SetStatus(codes.Ok, "File processed successfully")
	return recognitionResult, nil
}
