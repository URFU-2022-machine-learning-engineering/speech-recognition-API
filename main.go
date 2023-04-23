package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
)

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
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()

	// err = checkFileSignature(file)
	// if err != nil {
	// 	log.Println("File signature check failed:", err)
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
	fileName := handler.Filename
	fileExt := filepath.Ext(fileName)
	fileName = fmt.Sprintf("%s%s", generateRandomString(16), fileExt)
	
	err = uploadToMinio(fileName, file, handler.Size)
	if err != nil {
		log.Println("Failed to upload file to Minio:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("File uploaded successfully to Minio:", handler.Filename)

	// Send success response
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "File uploaded successfully")
}

func main() {
	// Set up HTTP server and routes
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/upload", uploadHandler)

	log.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
