package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestSendToProcess(t *testing.T) {
	// Store the original value of WHISPER_ENDPOINT and restore it at the end of the test
	originalWhisperEndpoint := os.Getenv("WHISPER_ENDPOINT")
	defer os.Setenv("WHISPER_ENDPOINT", originalWhisperEndpoint)

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
		w.Write(responseData)
	}))
	defer mockServer.Close()

	// Override the WHISPER_ENDPOINT with the mock server URL
	os.Setenv("WHISPER_ENDPOINT", mockServer.URL)

	// Call the handler function
	response, err := sendToProcess("my-bucket", "example.mp3")
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
