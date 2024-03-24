package utils

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"net/http"
	"sr-api/handlers/handlers_structure"
	"sr-api/helpers"
)

const errorClientMessage = "Failed to send message to recognition service"
const successClientMessage = "Sent to processing successfully, if I have any updates, I will let you know"

func ProcessFileToKafka(c *gin.Context, bucketName, fileName string) {
	_, span := helpers.StartSpanFromGinContext(c, "ProcessFileWithGinContext")
	defer span.End()

	log.Debug().Str("bucket", bucketName).Str("file", fileName).Msg("Initiating file processing")

	kafkaBroker, producerTopic, _ := helpers.GetKafkaConfig()

	data := handlers_structure.SendData{BucketName: bucketName, FileName: fileName}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Msg("Error marshaling data for Kafka")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error marshaling data for Kafka"})
		span.RecordError(err)
		return
	}
	writer := kafka.Writer{
		Addr:  kafka.TCP(kafkaBroker),
		Topic: producerTopic,
	}
	defer writer.Close()

	err = writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(fileName), // Optional: You could use the fileName as a key
			Value: jsonData,
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to send message to Kafka")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message to Kafka"})
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to send message to Kafka")
		return
	}

	log.Info().Str("file", fileName).Msg("Message sent to Kafka successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Message sent to Kafka successfully"})
	span.SetAttributes(attribute.String("kafka.message.status", "success"))
	span.SetStatus(codes.Ok, "Message sent to Kafka successfully")
}

// ProcessFileToKafkaWS processes a file information message and sends it to Kafka, using WebSocket for communication.
func ProcessFileToKafkaWS(conn *websocket.Conn, bucketName, fileName string) {
	// Context for tracing. Adjust as necessary for your application's tracing setup.
	ctx, span := helpers.StartSpanFromContext(context.Background(), "ProcessFileToKafkaWS")
	defer span.End()

	log.Debug().Str("bucket", bucketName).Str("file", fileName).Msg("Initiating file processing for WebSocket")

	kafkaBroker, producerTopic, _ := helpers.GetKafkaConfig()

	data := handlers_structure.SendData{BucketName: bucketName, FileName: fileName}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Msg("Error marshaling data for Kafka")
		respondWithErrorWS(conn)
		span.RecordError(err)
		return
	}

	writer := kafka.Writer{
		Addr:  kafka.TCP(kafkaBroker),
		Topic: producerTopic,
	}
	defer writer.Close()

	err = writer.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(fileName), // Optional: You could use the fileName as a key
			Value: jsonData,
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to send message to Kafka")
		respondWithErrorWS(conn)
		span.RecordError(err)
		return
	}

	log.Info().Str("file", fileName).Msg("Message sent to Kafka successfully")
	respondWithSuccessWS(conn)
}

// Helper function to send an error message over WebSocket.
func respondWithErrorWS(conn *websocket.Conn) {
	response := map[string]string{"error": errorClientMessage}
	if err := conn.WriteJSON(response); err != nil {
		log.Error().Err(err).Msg("Failed to send error response over WebSocket")
	}
}

// Helper function to send a success message over WebSocket.
func respondWithSuccessWS(conn *websocket.Conn) {
	response := map[string]string{"message": successClientMessage}
	if err := conn.WriteJSON(response); err != nil {
		log.Error().Err(err).Msg("Failed to send success response over WebSocket")
	}
}
