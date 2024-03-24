package utils

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"sr-api/handlers/handlers_structure"
	"sr-api/helpers"
)

// RespondWithError sends an error response along with tracing the operation.
func RespondWithError(c *gin.Context, code int, message string) {
	_, span := helpers.StartSpanFromGinContext(c, "RespondWithError")
	span.SetAttributes(attribute.Int("http.status_code", code), attribute.String("error.message", message))
	defer span.End()

	log.Error().Int("code", code).Str("message", message).Msg("Error response")
	span.SetStatus(codes.Error, message)

	c.JSON(code, gin.H{"error": message})
}

// RespondWithErrorOverWebSocket sends an error message over a WebSocket connection along with tracing the operation.
func RespondWithErrorOverWebSocket(span trace.Span, ws *websocket.Conn, code int, message string, err error) {

	span.SetAttributes(attribute.Int("websocket.status_code", code), attribute.String("error.message", message))
	defer span.End()

	log.Error().Err(err).Int("code", code).Str("message", message).Msg("WebSocket error response")
	span.RecordError(err)
	span.SetStatus(codes.Error, message)

	// Prepare the error response
	response := handlers_structure.WsErrorResponse{Error: message}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal error response")
		return
	}

	// Send the error message over the WebSocket connection
	if err := ws.WriteMessage(websocket.TextMessage, responseBytes); err != nil {
		log.Error().Err(err).Msg("Failed to send error response over WebSocket")
	}
}
