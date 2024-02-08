package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestProcessFileWithContext(t *testing.T) {
	// Store the original value of WHISPER_ENDPOINT and restore it at the end of the test
	ctx := context.Background()
	originalWhisperEndpoint := os.Getenv("WHISPER_ENDPOINT")
	originalTranscribeUri := os.Getenv("WHISPER_TRANSCRIBE")
	defer func() {
		err := os.Setenv("WHISPER_ENDPOINT", originalWhisperEndpoint)
		if err != nil {
			t.Errorf("Unable to set WHISPER_ENDPOINT ENV '%v", err)
		}
		err = os.Setenv("WHISPER_TRANSCRIBE", originalTranscribeUri)
		if err != nil {
			t.Errorf("Unable to set WHISPER_TRANSCRIBE ENV '%v", err)
		}
	}()

	// Mock the environment variable
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the HTTP method and request body
		if r.Method != http.MethodPost {
			t.Errorf("Expected 'POST' request, got '%s'", r.Method)
		}

		// Parse the expected request payload
		expectedData := map[string]interface{}{
			"file_name": "example.mp3",
			"bucket":    "my-bucket",
		}
		expectedJSONData, _ := json.Marshal(expectedData)

		body, _ := io.ReadAll(r.Body)
		if !bytes.Equal(body, expectedJSONData) {
			t.Errorf("Expected request body '%s', got '%s'", string(expectedJSONData), string(body))
		}

		// Send a mock response
		responseJSON := map[string]interface{}{
			"status": "success",
		}
		responseData, _ := json.Marshal(responseJSON)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(responseData)
		if err != nil {
			return
		}
	}))
	defer mockServer.Close()

	// Override the WHISPER_ENDPOINT with the mock server URL
	err := os.Setenv("WHISPER_ENDPOINT", mockServer.URL)
	if err != nil {
		t.Errorf("Unable to set WHISPER_ENDPOINT ENV '%v", err)
		return
	}
	err = os.Setenv("WHISPER_TRANSCRIBE", "/transcribe")
	if err != nil {
		t.Errorf("Unable to set WHISPER_TRANSCRIBE ENV '%v", err)
		return
	}

	// Call the handler function
	response, err := ProcessFileWithContext(ctx, "my-bucket", "example.mp3")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check the response data
	expectedResponse := map[string]interface{}{
		"status": "success",
	}
	if !reflect.DeepEqual(response, expectedResponse) {
		t.Errorf("Expected response '%v', got '%v'", expectedResponse, response)
	}
}
