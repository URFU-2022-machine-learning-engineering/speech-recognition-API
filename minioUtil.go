package main

import (
	"context"
	"mime/multipart"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// uploadToMinio uploads a file to Minio
func uploadToMinio(filename string, file multipart.File, size int64) error {

	// Get Minio access key and secret key from environment variables
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	minioBucket := os.Getenv("MINIO_BUCKET")

	// Set up Minio client
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	minioUseSSL := false
	// Initialize Minio client
	minioClient, err := minio.New(
		minioEndpoint,
		&minio.Options{
			Creds:  credentials.NewStaticV4(
				minioAccessKey,
				minioSecretKey,
				"",
			),
			Secure: minioUseSSL,
		},
	)
	if err != nil {
		return err
	}

	_, err = minioClient.PutObject(
		context.Background(),
		minioBucket,
		filename,
		file,
		size,
		minio.PutObjectOptions{},
	)
	if err != nil {
		return err
	}

	return nil
}
