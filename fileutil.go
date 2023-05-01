package main

import (
	"fmt"
	"mime/multipart"

	"github.com/h2non/filetype"
)

const signatureLength = 261

// checkFileSignature checks if the given file has a valid MP3 file signature.
// It returns an error if the file is not an audio file or if there was an error reading from the file.
func checkFileSignature(file multipart.File) (err error) {
	// Check file size before reading signature
	size, err := file.Seek(0, 2) // Seek to end of file to get size
	if err != nil {
		return err
	}
	if size < signatureLength {
		return fmt.Errorf("File is too small to be an audio file")
	}

	// Read signature bytes and check if they match an audio file signature
	_, err = file.Seek(0, 0) // Seek back to beginning of file
	if err != nil {
		return err
	}
	sig := make([]byte, signatureLength)
	_, err = file.Read(sig)
	if err != nil {
		return err
	}
	if !filetype.IsAudio(sig) {
		return fmt.Errorf("File is not an audio file")
	}

	// Reset file offset back to beginning of file
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	return
}
