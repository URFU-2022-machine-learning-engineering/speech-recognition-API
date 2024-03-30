package config

import (
	"strconv"
)

func LoadConfig() *AppConfig {
	minioUseSSL, err := strconv.ParseBool(GetEnv("MINIO_USE_SSL"))
	if err != nil {
		minioUseSSL = false
	}

	config := &AppConfig{
		MinioAccessKey:        GetEnv("MINIO_ACCESS_KEY"),
		MinioSecretKey:        GetEnv("MINIO_SECRET_KEY"),
		MinioEndpoint:         GetEnv("MINIO_ENDPOINT"),
		MinioBucket:           GetEnv("MINIO_BUCKET"),
		MinioUseSSL:           minioUseSSL,
		WhisperEndpoint:       GetEnv("WHISPER_ENDPOINT"),
		WhisperTranscribe:     GetEnv("WHISPER_TRANSCRIBE"),
		TelemetryGrpcEndpoint: GetEnv("TELEMETRY_GRPC_TARGET"),
	}

	return config
}
