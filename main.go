package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type PayloadSuccess struct {
	Result string
	Filename string
}

type PayloadError struct {
	Result string
}



func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		fmt.Fprintln(w, "Server is online")
	} else {
		// Return a 405 Method Not Allowed response for non-GET requests
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "Method Not Allowed")
	}
}


func uploadHandler(w http.ResponseWriter, r *http.Request) {
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Println("Failed to get uploaded file:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		p := PayloadError{"Failed to get uploaded file"}
		err := json.NewEncoder(w).Encode(p)
		if err != nil {
			return
		}
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)

	err = checkFileSignature(file)
	if err != nil {
		log.Println("File signature check failed:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		p := PayloadError{"Failed to upload file"}
		err := json.NewEncoder(w).Encode(p)
		if err != nil {
			return
		}
		return
	}

	fileExt := filepath.Ext(handler.Filename)
	fileName := fmt.Sprintf("%s%s", generateRandomString(16), fileExt)
	err = uploadToMinio(fileName, file, handler.Size)
	if err != nil {
		log.Println("Failed to upload file to Minio:", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		p := PayloadError{"Failed to upload file"}
		err := json.NewEncoder(w).Encode(p)
		if err != nil {
			return
		}
		return
	}

	log.Println("File uploaded successfully to Minio:", fileName)
	// Send success response
	res, err := sendToProcess(os.Getenv("MINIO_BUCKET"), fileName)
	if err != nil {
		log.Println("Field to transcribe file", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		return
	}
}

func main() {
	// Set up HTTP server and routes
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/upload", uploadHandler)

	log.Println("Server started at 0.0.0.0:8080")
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		return
	}
}

