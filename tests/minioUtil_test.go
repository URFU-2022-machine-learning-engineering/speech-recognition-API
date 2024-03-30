package tests

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/http/httptest"
	minio2 "sr-api/internal/adapters/repository"
	"sr-api/internal/config"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"

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

func createTestConfig(minioURL string) *config.AppConfig {
	return &config.AppConfig{
		MinioAccessKey: "minioadmin",
		MinioSecretKey: "minioadmin",
		MinioEndpoint:  minioURL,
		MinioBucket:    "test-bucket",
		MinioUseSSL:    false,
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
	// Existing setup for test containers...

	ctx := context.Background()

	// Start the MinIO Testcontainer
	minioClient, minioURL, cleanup := StartMinioTestContainer(ctx)
	defer cleanup()

	// Create a test configuration with the test MinIO container's URL
	testConfig := createTestConfig(minioURL)

	// Initialize MinioRepository with the test configuration
	minioRepo, err := minio2.NewMinioRepository(testConfig)
	if err != nil {
		t.Fatalf("Failed to create MinioRepository: %v", err)
	}
	if minioRepo == nil {
		t.Fatal("MinioRepository is nil")
	}
	err = minioClient.MakeBucket(ctx, testConfig.MinioBucket, minio.MakeBucketOptions{})
	if err != nil {
		t.Fatalf("Failed to create bucket: %v", err)
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/upload", func(c *gin.Context) {
		// Simulate file upload as before, using the mockAudioFile struct
		file := &mockAudioFile{content: "test content"}
		if err := minioRepo.UploadToMinioWithContext(c, "testfile.mp3", file, int64(len(file.content))); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
	})

	// Create a test request to the route
	req, _ := http.NewRequest(http.MethodPost, "/upload", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	if w.Code != http.StatusOK {
		t.Errorf("Expected HTTP status 200, got: %d", w.Code)
	}
}
