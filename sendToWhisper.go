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


func sendToProcess(bucketName string, fileName string) (map[string]interface{}, error) {
	whisper_endpoint := os.Getenv("WHISPER_ENDPOINT")
	whisper_transcribe := os.Getenv("WHISPER_TRANSCRIBE")
	whisperTranscribeURL := whisper_endpoint + path.Join(whisper_transcribe)

	log.Println("whisper url", whisperTranscribeURL)
	// 1. Marshal your data into JSON format
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

	// 6. Unmarshal the response JSON
	var responseJSON map[string]interface{}
	err = json.Unmarshal(body, &responseJSON)
	if err != nil {
		log.Panicln("Unmarshal the response JSON", err)
		return nil, err
	}
	log.Println("Received response:", responseJSON)
	return responseJSON, nil
}
