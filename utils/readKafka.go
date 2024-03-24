package utils

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"sr-api/helpers"
	"time"
)

// RecognitionResult represents the structure of messages in the 'recognition' topic.
type RecognitionResult struct {
	DetectedLanguage string `json:"detected_language"`
	RecognizedText   string `json:"recognized_text"`
}

// ReadFromKafkaToWebSocket reads messages from Kafka's 'recognition' topic and sends them over WebSocket,
// filtering messages by a specific key.
func ReadFromKafkaToWebSocket(conn *websocket.Conn, fileUUID string, userID string) {
	kafkaBroker, _, consumerTopic := helpers.GetKafkaConfig()
	log.Debug().Str("file_uuid", fileUUID).Str("user_id", userID).Msg("Initiating Kafka reader")
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaBroker},
		Topic:   consumerTopic,
		GroupID: userID,
	})
	defer reader.Close()

	// Set a 5-minute timeout
	waitTimeout := 5 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
	defer cancel()

	found := false
	for !found {
		log.Debug().Msg("Waiting for message from Kafka")
		msg, err := reader.FetchMessage(ctx)
		log.Debug().Msg("Fetching message from Kafka")
		if err != nil {
			// Check if the context deadline was exceeded, which means the timeout occurred
			if ctx.Err() == context.DeadlineExceeded {
				log.Error().Msg("Timeout reached waiting for message")
				respondWithErrorWS(conn)
				break
			}
			log.Error().Err(err).Msg("Failed to fetch message from Kafka")
			respondWithErrorWS(conn)
			break
		}

		if string(msg.Key) == fileUUID {
			log.Debug().Str("msg", string(msg.Value)).Msg("Received message from Kafka")

			var result RecognitionResult
			if err := json.Unmarshal(msg.Value, &result); err != nil {
				log.Error().Err(err).Msg("Error unmarshalling recognition result")
				respondWithErrorWS(conn)
				continue
			}

			log.Info().Str("detected_language", result.DetectedLanguage).Str("recognized_text", result.RecognizedText).Msg("Processed recognition result")

			if err := conn.WriteJSON(result); err != nil {
				log.Error().Err(err).Msg("Failed to send recognition result over WebSocket")
			} else {
				found = true // Stop the loop if the message is successfully sent
			}
		} else {
			// Commit the message to advance the consumer's offset.
			// This is important to not read the same message again and again.
			if err := reader.CommitMessages(ctx, msg); err != nil {
				log.Error().Err(err).Msg("Failed to commit message")
			}
		}
	}

	if !found {
		// Handle the case where the message was not found within the timeout
		// Maybe send a specific message back over WebSocket
		log.Error().Msg("Message not found within timeout")
		respondWithErrorWS(conn)
	}
}
