package main

import (
	"context"
	"log"
	"mime/multipart"
	"os"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// uploadToMinio uploads a file to Minio
func uploadToMinio(filename string, file multipart.File, size int64) error {
	// Load environment variables from .env.local file
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatal("Error loading .env.local file:", err)
	}

	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	minioBucket := os.Getenv("MINIO_BUCKET")
	minioUseSSL := false

	// Initialize Minio client
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
			Secure: minioUseSSL,
		},
	)
	if err != nil {
		log.Fatal(err)
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
