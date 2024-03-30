package ports

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/codes"
	"net/http"
	"sr-api/internal/core/ports/telemetry"
)

func HttpClient(c *gin.Context, request *http.Request) (*http.Response, error) {
	_, span := telemetry.StartSpanFromGinContext(c, "SendRequestToWhisperService")
	spanID := telemetry.GetSpanId(span)
	log.Debug().Str("span_id", spanID).Msg("Starting httpClient")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Error().Str("span_id", spanID).Err(err).Msg("Failed to send request to Whisper service")
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to send request to Whisper service")
		RespondWithError(c, span, err, http.StatusInternalServerError, "Failed to send request to Whisper service")
		return nil, err
	}
	log.Info().Str("span_id", spanID).Msg("Request sent to Whisper service")
	if response.StatusCode != http.StatusOK {
		log.Error().Str("span_id", spanID).Int("status_code", response.StatusCode).Msg("Whisper service returned an error")
		span.SetStatus(codes.Error, fmt.Sprintf("Whisper service error: %d", response.StatusCode))
		RespondWithError(c, span, err, response.StatusCode, fmt.Sprintf("Whisper service error: %d", response.StatusCode))
		return nil, fmt.Errorf("whisper service error: %d", response.StatusCode)
	}
	return response, nil
}
