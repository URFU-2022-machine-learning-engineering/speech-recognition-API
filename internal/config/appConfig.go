package config

type AppConfig struct {
	MinioAccessKey        string
	MinioSecretKey        string
	MinioEndpoint         string
	MinioBucket           string
	MinioUseSSL           bool
	WhisperEndpoint       string
	WhisperTranscribe     string
	TelemetryGrpcEndpoint string
}
