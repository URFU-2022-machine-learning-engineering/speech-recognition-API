package helpers

import (
	"github.com/rs/zerolog/log"
	"os"
)

// GetMinIOConfig is a function that retrieves the configuration for MinIO from environment variables.
// It returns five strings representing the access key, secret key, endpoint, bucket, and SSL status.
//
// The function checks if the following environment variables are set:
// - MINIO_ACCESS_KEY: The access key for MinIO. If it's not set, the function logs a fatal error.
// - MINIO_SECRET_KEY: The secret key for MinIO. If it's not set, the function logs a fatal error.
// - MINIO_ENDPOINT: The endpoint for MinIO. If it's not set, the function logs a fatal error.
// - MINIO_BUCKET: The bucket for MinIO. If it's not set, the function logs a fatal error.
// - MINIO_SSL: The SSL status for MinIO. If it's not set, the function logs a warning and defaults to false.
//
// The function returns the values of these environment variables in the order they are listed above.
func GetMinIOConfig() (string, string, string, string, string) {

	mak := os.Getenv("MINIO_ACCESS_KEY")
	if mak == "" {
		log.Fatal().Msg("MINIO_ACCESS_KEY environment variable is not set")
	}
	msk := os.Getenv("MINIO_SECRET_KEY")
	if msk == "" {
		log.Fatal().Msg("MINIO_SECRET_KEY environment variable is not set")
	}
	me := os.Getenv("MINIO_ENDPOINT")
	if me == "" {
		log.Fatal().Msg("MINIO_ENDPOINT environment variable is not set")
	}
	mb := os.Getenv("MINIO_BUCKET")
	if mb == "" {
		log.Fatal().Msg("MINIO_BUCKET environment variable is not set")
	}
	mssl := os.Getenv("MINIO_SSL")
	if mssl == "" {
		log.Warn().Msg("MINIO_SSL environment variable is not set, defaulting to false")
	}

	return mak, msk, me, mb, mssl
}

func GetKafkaConfig() (string, string, string) {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		log.Fatal().Msg("KAFKA_BROKER environment variable is not set")
	}
	producerTopic := os.Getenv("PRODUCER_KAFKA_TOPIC")
	if producerTopic == "" {
		log.Fatal().Msg("PRODUCER_KAFKA_TOPIC environment variable is not set")
	}

	consumerTopic := os.Getenv("CONSUMER_KAFKA_TOPIC")
	if producerTopic == "" {
		log.Fatal().Msg("CONSUMER_KAFKA_TOPIC environment variable is not set")
	}

	return broker, producerTopic, consumerTopic
}
