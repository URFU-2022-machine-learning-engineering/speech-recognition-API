package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

type Payload struct {
	Result resultValue
}

type resultValue string

// Define root handler
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		fmt.Fprintln(w, "Server is online")
	} else {
		// Return a 405 Method Not Allowed response for non-GET requests
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "Method Not Allowed")
	}
}

// uploadHandler handles file uploads
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Println("Failed to get uploaded file:", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		p := Payload{"Failed to get uploaded file"}
		err := json.NewEncoder(w).Encode(p)
		if err != nil {
			return
		}
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	err = checkFileSignature(file)
	if err != nil {
		log.Println("File signature check failed:", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		p := Payload{"Failed to upload file"}
		err := json.NewEncoder(w).Encode(p)
		if err != nil {
			return
		}
		return
	}

	fileName := handler.Filename
	fileExt := filepath.Ext(fileName)
	fileName = fmt.Sprintf("%s%s", generateRandomString(16), fileExt)

	err = uploadToMinio(fileName, file, handler.Size)
	if err != nil {
		log.Println("Failed to upload file to Minio:", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		p := Payload{"Failed to upload file"}
		err := json.NewEncoder(w).Encode(p)
		if err != nil {
			return
		}
		return
	}

	log.Println("File uploaded successfully to Minio:", handler.Filename)
	// Send success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	p := Payload{"File uploaded successfully"}
	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		return
	}
}

func main() {
	// Set up HTTP server and routes
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/upload", uploadHandler)

	log.Println("Server started at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}
