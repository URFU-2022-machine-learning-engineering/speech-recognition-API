package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sr-api/handlers/handlers_structure"
	"strings"
	"testing"
)

func TestStatusHandlerPositive(t *testing.T) {
	// Create a new request with the GET method to the root path ("/")
	req, err := http.NewRequest("GET", "/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new ResponseRecorder (which satisfies http.ResponseWriter)
	rr := httptest.NewRecorder()

	// Call the rootHandler function with the request and response recorder
	handler := http.HandlerFunc(StatusHandler)
	handler.ServeHTTP(rr, req)

	// Check the status code returned by the handler
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body returned by the handler
	expected := handlers_structure.StatusSuccess{Status: "Server is running"}
	expectedJson, _ := json.Marshal(expected)
	expectedStr := strings.TrimSpace(string(expectedJson))
	actualStr := strings.TrimSpace(rr.Body.String())
	if actualStr != expectedStr {
		t.Errorf("handler returned unexpected body: got %v want %v",
			actualStr, expectedStr)
	}
}

func TestStatusHandlerNegative(t *testing.T) {
	// Create a new request with the POST method to the root path ("/")
	req, err := http.NewRequest("POST", "/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new ResponseRecorder (which satisfies http.ResponseWriter)
	rr := httptest.NewRecorder()

	// Call the rootHandler function with the request and response recorder
	handler := http.HandlerFunc(StatusHandler)
	handler.ServeHTTP(rr, req)

	// Check the status code returned by the handler
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}

	// Check the response body returned by the handler
	expected := "Method Not Allowed\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
