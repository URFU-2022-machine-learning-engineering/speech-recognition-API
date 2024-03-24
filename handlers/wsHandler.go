package handlers

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"os"
	"sr-api/helpers"
	"sr-api/utils"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Adjust this to a more secure version in production
	},
}

func WsHandler(c *gin.Context) {
	_, span := helpers.StartSpanFromGinContext(c, "WsHandler")
	defer span.End()

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade to websocket")
		return
	}
	defer func() {
		// Send a close message to the client indicating a normal closure
		msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Normal closure")
		if writeErr := ws.WriteMessage(websocket.CloseMessage, msg); writeErr != nil {
			log.Error().Err(writeErr).Msg("Failed to send close message")
		}
		ws.Close()
	}()
	userId := c.GetHeader("User-Id")
	if userId == "" {
		log.Error().Msg("No User-Id header provided")
		return
	}
	log.Info().Str("userId", userId).Msgf("Working with user %s", userId)

	// Listen on the new WebSocket connection
	for {
		// Reading message (expecting audio data in bytes)
		mt, message, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Msg("Error reading websocket message")
			}
			break
		}

		if mt == websocket.CloseMessage {
			log.Info().Msg("Received close message")
			break
		}
		log.Info().Msgf("Received message type: %v", mt)

		log.Info().Msgf("Received: %v bytes", len(message))
		// Convert message bytes to reader
		reader := bytes.NewReader(message)
		contentType, err := helpers.CheckFileSignatureWithGinContext(c, reader)
		if err != nil {
			log.Error().Err(err).Msg("Invalid file signature")
			utils.RespondWithErrorOverWebSocket(span, ws, websocket.CloseUnsupportedData, "Invalid file signature", err)
			break
		}
		fileUUID, err := helpers.GenerateUIDWithContext(c)
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate UUID for file")
			utils.RespondWithErrorOverWebSocket(span, ws, websocket.CloseInternalServerErr, "Failed to process file", err)
			break
		}

		if err := utils.UploadToMinioWithContext(c, fileUUID, reader, int64(len(message))); err != nil {
			utils.RespondWithErrorOverWebSocket(span, ws, websocket.CloseInternalServerErr, "Failed to process file", err)
			break
		}

		log.Info().Str("fileName", fileUUID).Str("content_type", contentType).Msg("File uploaded successfully")
		span.AddEvent("File uploaded successfully", trace.WithAttributes(attribute.String("fileName", fileUUID)))

		// Process the file after successful upload
		utils.ProcessFileToKafkaWS(ws, os.Getenv("MINIO_BUCKET"), fileUUID)
		log.Debug().Str("bucket", os.Getenv("MINIO_BUCKET")).Str("file_name", fileUUID).Msg("Initiated file processing")
		span.SetStatus(codes.Ok, "Successfully uploaded the file")

		// After processing the audio data, you can respond with JSON data
		// Replace the following with your actual response logic
		log.Debug().Msg("Waiting for audio processing to complete")
		utils.ReadFromKafkaToWebSocket(ws, fileUUID, userId)
		response := map[string]string{"status": "success", "message": "Audio processed"}
		if err := ws.WriteJSON(response); err != nil {
			log.Error().Err(err).Msg("Error sending response")
			break
		}

		//Optional: If you're only expecting a single message per connection, break the loop here
		break
	}
}
