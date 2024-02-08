package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

type SendData struct {
	BucketName string `json:"bucket"`
	FileName   string `json:"file_name"`
}

// processResponse is a function that takes a byte slice as input and attempts to unmarshal it as JSON.
// If the unmarshalling is successful, it logs the JSON response and returns it.
// If the unmarshalling fails, it treats the input as an HTML response, logs it, and returns it.
// The function returns two values: the response (either JSON or HTML) and an error (if any occurred during the process).
func processResponse(body []byte) (interface{}, error) {
	// Declare a variable of type interface{} to hold the response
	var response interface{}

	// Attempt to unmarshal the response as JSON
	err := json.Unmarshal(body, &response)
	if err == nil {
		// If the unmarshalling is successful, log the JSON response and return it
		log.Println("Received JSON response:", response)
		return response, nil
	}

	// If unmarshalling as JSON fails, treat it as an HTML response
	htmlResponse := string(body)
	// Log the HTML response
	log.Println("Received HTML response:", htmlResponse)
	// Return the HTML response
	return htmlResponse, nil
}

func ProcessFileWithContext(ctx context.Context, bucketName, fileName string) (interface{}, error) {
	tracer := otel.Tracer("utils")
	_, span := tracer.Start(ctx, "ProcessFileWithContext")
	defer span.End()

	whisperEndpoint := os.Getenv("WHISPER_ENDPOINT")
	whisperTranscribe := os.Getenv("WHISPER_TRANSCRIBE")
	whisperTranscribeURL := path.Join(whisperEndpoint, whisperTranscribe)

	data := SendData{BucketName: bucketName, FileName: fileName}
	jsonData, err := json.Marshal(data)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", whisperTranscribeURL, bytes.NewBuffer(jsonData))
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			span.RecordError(err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	response, err := processResponse(body)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("response.error", "Failed to process response"))
		return nil, err
	}

	span.SetAttributes(attribute.String("response.type", "Success"))
	return response, nil
}
