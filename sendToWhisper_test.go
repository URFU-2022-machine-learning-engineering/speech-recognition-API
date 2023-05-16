package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)


func TestSendToProcess(t *testing.T) {
	// Mock the environment variable
	os.Setenv("WHISPER_ENDPOINT", "http://example.com/whisper")

	// Create a test request payload
	fileName := "example.mp3"
	bucketName := "my-bucket"
	expectedData := map[string]interface{}{
		"file_name":   fileName,
		"bucket": bucketName,
	}
	jsonData, _ := json.Marshal(expectedData)

	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the HTTP method and request body
		if r.Method != "POST" {
			t.Errorf("Expected 'POST' request, got '%s'", r.Method)
		}

		body, _ := ioutil.ReadAll(r.Body)
		if !bytes.Equal(body, jsonData) {
			t.Errorf("Expected request body '%s', got '%s'", string(jsonData), string(body))
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
	response, err := sendToProcess(bucketName, fileName)

	// Check for errors
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


