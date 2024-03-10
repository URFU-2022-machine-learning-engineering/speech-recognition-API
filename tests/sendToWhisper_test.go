package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sr-api/utils"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestProcessFileWithGinContext(t *testing.T) {
	// Setup Gin
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Mock server for handling external service calls
	mockServer := setupMockServer()
	defer mockServer.Close()

	// Set environment variables for the duration of the test
	setTestEnvVariables(mockServer.URL)
	defer restoreEnvVariables()

	// Define the endpoint within your test where `ProcessFileWithGinContext` is utilized
	r.POST("/test", func(c *gin.Context) {
		utils.ProcessFileWithGinContext(c, "my-bucket", "example.mp3")
	})

	// Create a test request and recorder
	req, _ := http.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert the HTTP status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Assert the response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}

	expectedResponse := map[string]interface{}{
		"detected_language": "en",
		"recognized_text":   "sample text",
	}
	if !reflect.DeepEqual(response, expectedResponse) {
		t.Errorf("Expected response %+v, got %+v", expectedResponse, response)
	}
}

func setupMockServer() *httptest.Server {
	// Initialize and return a mock server that simulates the external service
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a JSON response that includes detected_language and recognized_text.
		mockResponse := map[string]interface{}{
			"detected_language": "en",
			"recognized_text":   "sample text",
		}
		responseData, _ := json.Marshal(mockResponse)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseData)
	}))
	return mockServer
}

func setTestEnvVariables(mockServerURL string) {
	// Set environment variables to use the mock server and any other necessary configurations
	os.Setenv("WHISPER_ENDPOINT", mockServerURL)
	os.Setenv("WHISPER_TRANSCRIBE", "/transcribe")
	os.Setenv("WHISPER_API_KEY", "test")
}

func restoreEnvVariables() {
	// Restore or clear environment variables set by setTestEnvVariables
	os.Unsetenv("WHISPER_ENDPOINT")
	os.Unsetenv("WHISPER_TRANSCRIBE")
	os.Unsetenv("WHISPER_API_KEY")
}
