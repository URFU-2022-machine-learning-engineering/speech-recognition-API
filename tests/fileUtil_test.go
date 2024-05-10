package tests

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httptest"
	"sr-api/internal/core/domain"
	"strings"
	"testing"
)

// MockFile Define a mock implementation of multipart.File for testing purposes
type MockFile struct {
	content string
	offset  int64
}

func (m *MockFile) Seek(offset int64, whence int) (int64, error) {
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

func (m *MockFile) Read(p []byte) (n int, err error) {
	if m.offset >= int64(len(m.content)) {
		return 0, io.EOF
	}
	n = copy(p, m.content[m.offset:])
	m.offset += int64(n)
	return n, nil
}

func (m *MockFile) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(m.content)) {
		return 0, io.EOF
	}
	n = copy(p, m.content[off:])
	return n, nil
}

func (m *MockFile) Close() error {
	return nil // Mock implementation, does nothing
}

func performFileSignatureCheckTest(t *testing.T, fileContent string, expectedStatus int, expectedBody string) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	file := &MockFile{content: fileContent}

	r.POST("/test-file-signature", func(c *gin.Context) {
		err := domain.CheckFileSignatureWithGinContext(c, file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "File signature verification successful"})
	})

	req, _ := http.NewRequest("POST", "/test-file-signature", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != expectedStatus {
		t.Errorf("Expected status %d, got: %d, body: %s", expectedStatus, w.Code, w.Body.String())
	}

	if expectedBody != "" && !strings.Contains(w.Body.String(), expectedBody) {
		t.Errorf("Expected body to contain '%s', got: '%s'", expectedBody, w.Body.String())
	}
}

func TestCheckFileSignature_ValidAudioFile(t *testing.T) {
	validFileContent := "\xFF\xFB" + strings.Repeat("\x00", domain.SignatureLength-2) // Simulate a valid audio file signature
	performFileSignatureCheckTest(t, validFileContent, http.StatusOK, "File signature verification successful")
}

func TestCheckFileSignature_SmallFile(t *testing.T) {
	smallFileContent := strings.Repeat("\x00", domain.SignatureLength-1) // Simulate a file that's too small
	performFileSignatureCheckTest(t, smallFileContent, http.StatusBadRequest, "{\"error\":\"file size is too small\"}")
}

func TestCheckFileSignature_NonAudioFile(t *testing.T) {
	nonAudioFileContent := "Hello, World!" + strings.Repeat("\x00", domain.SignatureLength-12) // Simulate a non-audio file
	performFileSignatureCheckTest(t, nonAudioFileContent, http.StatusBadRequest, "unknown file type")
}
