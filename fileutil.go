package main

import (
	"fmt"
	"mime/multipart"

	"github.com/h2non/filetype"
)

// checkFileSignature checks if the given file has a valid MP3 file signature
func checkFileSignature(file multipart.File) error {
	buf := make([]byte, 261) // Read first 261 bytes for signature check
	_, err := file.Read(buf)
	if err != nil {
		return err
	}
	kind, err := filetype.Match(buf)
	if err != nil {
		return err
	}

	if kind.MIME.Value != "audio/mpeg" {
		return fmt.Errorf("File is not an MP3 file")
	}

	return nil
}