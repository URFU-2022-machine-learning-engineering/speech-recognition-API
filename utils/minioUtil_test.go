package utils

import (
	"context"
	"log"
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
	os.Setenv("MINIO_ACCESS_KEY", originalAccessKey)
	os.Setenv("MINIO_SECRET_KEY", originalSecretKey)
	os.Setenv("MINIO_BUCKET", originalBucket)
	os.Setenv("MINIO_USE_SSL", originalUseSSL)
	os.Setenv("MINIO_ROOT_USER", originalRootUser)
	os.Setenv("MINIO_ROOT_PASSWORD", originalRootPassword)
}

func setupEnvs() {
	// Set the necessary environment variables for the test
	os.Setenv("MINIO_ACCESS_KEY", "minioadmin")
	os.Setenv("MINIO_BUCKET", "test-bucket")
	os.Setenv("MINIO_SECRET_KEY", "minioadmin")
	os.Setenv("MINIO_USE_SSL", "false")
	os.Setenv("MINIO_ROOT_USER", "minioadmin")
	os.Setenv("MINIO_ROOT_PASSWORD", "minioadmin")

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
	enp, err := minioContainer.Endpoint(ctx, "")
	log.Println(enp)

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
	ac := os.Getenv("MINIO_ACCESS_KEY")
	log.Println(ac)

	ctx := context.Background()
	// Inside your test or helper function

	// Start the MinIO Testcontainer
	minioClient, minioURL, cleanup := StartMinioTestContainer(ctx)
	defer cleanup()

	// Set MINIO_ENDPOINT environment variable for the test
	originalEndpoint := os.Getenv("MINIO_ENDPOINT")
	defer os.Setenv("MINIO_ENDPOINT", originalEndpoint) // Reset after the test
	os.Setenv("MINIO_ENDPOINT", minioURL)

	bucketName := "test-bucket"
	err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		t.Fatalf("Failed to create bucket: %v", err)
	}

	file := &mockAudioFile{content: "test content"}
	err = UploadToMinioWithContext(ctx, "testfile", file, int64(len(file.content)))
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}
