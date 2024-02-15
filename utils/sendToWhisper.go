package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"sr-api/handlers/handlers_structure"
)

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
	ctx, span := tracer.Start(ctx, "ProcessFileWithContext")
	defer span.End()
	span.AddEvent("Get the environment variables")
	whisperEndpoint := GetEnvOrShutdownWithTelemetry(ctx, "WHISPER_ENDPOINT")
	whisperTranscribe := GetEnvOrShutdownWithTelemetry(ctx, "WHISPER_TRANSCRIBE")
	span.AddEvent("Finished getting the environment variables")
	// Parse the base URL
	span.AddEvent("Parse the base URL")
	u, err := url.Parse(whisperEndpoint)
	if err != nil {
		span.RecordError(err)
		log.Fatalf("Failed to parse WHISPER_ENDPOINT: %v", err)
	}
	span.AddEvent("Finished parsing the base URL")
	// Properly append the path
	u.Path = path.Join(u.Path, whisperTranscribe)
	whisperTranscribeURL := u.String()
	span.AddEvent("Prepare the request")
	data := handlers_structure.SendData{BucketName: bucketName, FileName: fileName}
	jsonData, err := json.Marshal(data)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("Prepare the request")
	req, err := http.NewRequestWithContext(ctx, "POST", whisperTranscribeURL, bytes.NewBuffer(jsonData))
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	span.AddEvent("Send the request", trace.WithAttributes(attribute.String("request.type", "POST")))
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
	span.AddEvent("Read the file")
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
