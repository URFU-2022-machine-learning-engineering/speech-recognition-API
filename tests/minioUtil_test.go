package tests

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/http/httptest"
	"sr-api/utils"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"

	"os"
	"testing"
)

type mockAudioFile struct {
	content string
	offset  int64
}

func (m *mockAudioFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		m.offset = offset
	case io.SeekCurrent:
		m.offset += offset
	case io.SeekEnd:
		m.offset = int64(len(m.content)) + offset
	}
	return m.offset, nil
}

func (m *mockAudioFile) Read(p []byte) (n int, err error) {
	if m.offset >= int64(len(m.content)) {
		return 0, io.EOF
	}
	n = copy(p, m.content[m.offset:])
	m.offset += int64(n)
	return n, nil
}

func (m *mockAudioFile) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(m.content)) {
		return 0, io.EOF
	}
	n = copy(p, m.content[off:])
	return n, nil
}

func (m *mockAudioFile) Close() error {
	return nil
}

func restoreEnvs(
	originalAccessKey string,
	originalSecretKey string,
	originalBucket string,
	originalUseSSL string,
	originalRootUser string,
	originalRootPassword string,
) {
	err := os.Setenv("MINIO_ACCESS_KEY", originalAccessKey)
	if err != nil {
		log.Fatalf("Failed to restore original environment variable: %v", err)
	}
	err = os.Setenv("MINIO_SECRET_KEY", originalSecretKey)
	if err != nil {
		log.Fatalf("Failed to restore original environment variable: %v", err)
	}
	err = os.Setenv("MINIO_BUCKET", originalBucket)
	if err != nil {
		log.Fatalf("Failed to restore original environment variable: %v", err)
	}
	err = os.Setenv("MINIO_USE_SSL", originalUseSSL)
	if err != nil {
		log.Fatalf("Failed to restore original environment variable: %v", err)
	}
	err = os.Setenv("MINIO_ROOT_USER", originalRootUser)
	if err != nil {
		log.Fatalf("Failed to restore original environment variable: %v", err)
	}
	err = os.Setenv("MINIO_ROOT_PASSWORD", originalRootPassword)
	if err != nil {
		log.Fatalf("Failed to restore original environment variable: %v", err)
	}
}

func setupEnvs() {
	// Set the necessary environment variables for the test
	err := os.Setenv("MINIO_ACCESS_KEY", "minioadmin")
	if err != nil {
		log.Fatalf("Failed to set environment variable: %v", err)
	}
	err = os.Setenv("MINIO_BUCKET", "test-bucket")
	if err != nil {
		log.Fatalf("Failed to set environment variable: %v", err)
	}
	err = os.Setenv("MINIO_SECRET_KEY", "minioadmin")
	if err != nil {
		log.Fatalf("Failed to set environment variable: %v", err)
	}
	err = os.Setenv("MINIO_USE_SSL", "false")
	if err != nil {
		log.Fatalf("Failed to set environment variable: %v", err)
	}
	err = os.Setenv("MINIO_ROOT_USER", "minioadmin")
	if err != nil {
		log.Fatalf("Failed to set environment variable: %v", err)
	}
	err = os.Setenv("MINIO_ROOT_PASSWORD", "minioadmin")
	if err != nil {
		log.Fatalf("Failed to set environment variable: %v", err)
	}

}

// StartMinioTestContainer starts a MinIO server in a Docker container for testing.
func StartMinioTestContainer(ctx context.Context) (*minio.Client, string, func()) {
	req := testcontainers.ContainerRequest{
		Image:        "minio/minio",
		ExposedPorts: []string{"9000/tcp"},
		Env: map[string]string{
			"MINIO_ROOT_USER":     "minioadmin",
			"MINIO_ROOT_PASSWORD": "minioadmin",
		},
		Cmd:        []string{"server", "/data"},
		WaitingFor: wait.ForLog("Status:         1 Online, 0 Offline.").WithStartupTimeout(time.Second * 120),
	}

	minioContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Failed to start MinIO Testcontainer: %v", err)
	}

	minioURL, err := minioContainer.Endpoint(ctx, "")

	// Initialize MinIO client with NewStatic using SignatureV4
	minioClient, err := minio.New(minioURL, &minio.Options{
		Creds:  credentials.NewStatic("minioadmin", "minioadmin", "", credentials.SignatureV4),
		Secure: false,
	})

	if err != nil {
		log.Fatalf("Failed to create MinIO client: %v", err)
	}

	// Function to stop the container
	cleanup := func() {
		if err := minioContainer.Terminate(ctx); err != nil {
			log.Printf("Failed to terminate MinIO Testcontainer: %v", err)
		}
	}

	return minioClient, minioURL, cleanup
}

func TestUploadToMinioWithTestContainer(t *testing.T) {
	originalAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	originalSecretKey := os.Getenv("MINIO_SECRET_KEY")
	originalBucket := os.Getenv("MINIO_BUCKET")
	originalUseSSL := os.Getenv("MINIO_USE_SSL")
	originalRootUser := os.Getenv("MINIO_ROOT_USER")
	originalRootPassword := os.Getenv("MINIO_ROOT_PASSWORD")

	defer restoreEnvs(originalAccessKey, originalSecretKey, originalBucket, originalUseSSL, originalRootUser, originalRootPassword)
	setupEnvs()

	ctx := context.Background()

	// Start the MinIO Testcontainer
	minioClient, minioURL, cleanup := StartMinioTestContainer(ctx)
	defer cleanup()

	// Set MINIO_ENDPOINT environment variable for the test
	originalEndpoint := os.Getenv("MINIO_ENDPOINT")
	// Reset after the test
	defer func() {
		err := os.Setenv("MINIO_ENDPOINT", originalEndpoint)
		if err != nil {
			t.Fatalf("Failed to reset environment variable: %v", err)
		}
	}()

	err := os.Setenv("MINIO_ENDPOINT", minioURL)
	if err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}

	bucketName := "test-bucket"
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		t.Fatalf("Failed to create bucket: %v", err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Define a test route that uses UploadToMinioWithContext
	r.POST("/upload", func(c *gin.Context) {
		// Simulate receiving a file part as `file`, similar to how you'd receive it in a real request
		file := &mockAudioFile{content: "test content"}
		// Use the c.Request.Context() for operations that need a context.Context
		if err := utils.UploadToMinioWithContext(c, "testfile.mp3", file, int64(len(file.content))); err != nil {
			// Handle error...
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
	})

	// Create a test request to the route
	req, _ := http.NewRequest(http.MethodPost, "/upload", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check the response
	if w.Code != http.StatusOK {
		t.Errorf("Expected HTTP status 200, got: %d", w.Code)
	}
}
