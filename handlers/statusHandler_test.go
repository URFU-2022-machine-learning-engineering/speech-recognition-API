package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sr-api/handlers/handlers_structure"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestStatusHandlerPositive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Register your StatusHandler with the router
	router.GET("/status", StatusHandler)

	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Record the HTTP response using httptest
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check the status code is what we expect
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect
	expected := handlers_structure.StatusSuccess{Status: "ok"}
	expectedJson, _ := json.Marshal(expected)
	expectedStr := strings.TrimSpace(string(expectedJson))
	actualStr := strings.TrimSpace(rr.Body.String())
	if actualStr != expectedStr {
		t.Errorf("handler returned unexpected body: got %v want %v", actualStr, expectedStr)
	}
}
