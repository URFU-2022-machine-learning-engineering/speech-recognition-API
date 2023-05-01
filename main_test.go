package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRootHandlerPositive(t *testing.T) {
	// Create a new request with the GET method to the root path ("/")
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new ResponseRecorder (which satisfies http.ResponseWriter)
	rr := httptest.NewRecorder()

	// Call the rootHandler function with the request and response recorder
	handler := http.HandlerFunc(rootHandler)
	handler.ServeHTTP(rr, req)

	// Check the status code returned by the handler
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body returned by the handler
	expected := "Server is online\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestRootHandlerNegative(t *testing.T) {
	// Create a new request with the POST method to the root path ("/")
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new ResponseRecorder (which satisfies http.ResponseWriter)
	rr := httptest.NewRecorder()

	// Call the rootHandler function with the request and response recorder
	handler := http.HandlerFunc(rootHandler)
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
