package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"

	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7"
	"github.com/joho/godotenv"
)


func main() {
	// Load environment variables from .env.local file
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatal("Error loading .env.local file:", err)
	}

	// Get Minio access key and secret key from environment variables
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	minioBucket := os.Getenv("MINIO_BUCKET")

	// Set up Minio client
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	minioUseSSL := false



	// Create a new Minio client
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: minioUseSSL,
	})
	if err != nil {
		log.Fatal(err)
	}

		// Define root handler
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" {
				fmt.Fprint(w, "Online")
			} else {
				// Return a 405 Method Not Allowed response for non-GET requests
				w.WriteHeader(http.StatusMethodNotAllowed)
				fmt.Fprint(w, "Method Not Allowed")
			}
		})

	// Define the HTTP endpoint to receive the audio file from the frontend
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		// Parse the multipart form
		err := r.ParseMultipartForm(32 << 20) // 32 MB
		if err != nil {
			http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
			return
		}

		// Get the uploaded file from the frontend
		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to get uploaded file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Create a unique file name to store in Minio
		fileName := handler.Filename
		fileExt := filepath.Ext(fileName)
		fileName = fmt.Sprintf("%s%s", generateRandomString(16), fileExt)

		// Upload the file to Minio
		_, err = minioClient.PutObject(r.Context(), minioBucket, fileName, file, handler.Size, minio.PutObjectOptions{
			ContentType: handler.Header.Get("Content-Type"),
		})
		if err != nil {
			http.Error(w, "Failed to upload file to Minio", http.StatusInternalServerError)
			return
		}

		// Return the uploaded file URL to the frontend
		fileURL := fmt.Sprintf("http://%s/%s/%s", minioEndpoint, minioBucket, fileName)
		w.Write([]byte(fileURL))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Helper function to generate a random string
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
