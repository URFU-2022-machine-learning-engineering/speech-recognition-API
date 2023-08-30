package main

import (
	"bytes"
	"encoding/json"
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

func processResponse(body []byte) (interface{}, error) {
	var response interface{}

	// Attempt to unmarshal the response as JSON
	err := json.Unmarshal(body, &response)
	if err == nil {
		log.Println("Received JSON response:", response)
		return response, nil
	}

	// If unmarshaling as JSON fails, treat it as an HTML response
	htmlResponse := string(body)
	log.Println("Received HTML response:", htmlResponse)
	return htmlResponse, nil
}

func sendToProcess(bucketName string, fileName string) (interface{}, error) {
	whisperEndpoint := os.Getenv("WHISPER_ENDPOINT")
	whisperTranscribe := os.Getenv("WHISPER_TRANSCRIBE")
	whisperTranscribeURL := whisperEndpoint + path.Join(whisperTranscribe)

	log.Println("whisper url", whisperTranscribeURL)
	// 1. Marshal data into JSON format
	data := SendData{bucketName, fileName}
	log.Println("prepared", data)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// 2. Create an HTTP request
	req, err := http.NewRequest("POST", whisperTranscribeURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Panicln("Create HTTP", err)
		return nil, err
	}

	// 3. Set headers
	req.Header.Set("Content-Type", "application/json")
	log.Println("Request prepared", req)

	// 4. Send the request and handle the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panicln("Send the request and handle the response", err)
		return nil, err
	}
	defer resp.Body.Close()

	// 5. Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panicln("Read the response body", err)
		return nil, err
	}
	log.Println("Request was", body)

	// 6. Process the response
	return processResponse(body)
}
