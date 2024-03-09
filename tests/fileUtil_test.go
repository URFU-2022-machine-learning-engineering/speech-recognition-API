package tests

import (
	"context"
	"io"
	"sr-api/helpers"
	"strings"
	"testing"
)

func (m *mockFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		m.offset = offset
	case io.SeekCurrent:
		m.offset += offset
	case io.SeekEnd:
		m.offset = int64(len(m.content)) + offset
	}
	return m.offset, nil
}

func (m *mockFile) Read(p []byte) (n int, err error) {
	if m.offset >= int64(len(m.content)) {
		return 0, io.EOF
	}
	n = copy(p, m.content[m.offset:])
	m.offset += int64(n)
	return n, nil
}

func (m *mockFile) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(m.content)) {
		return 0, io.EOF
	}
	n = copy(p, m.content[off:])
	return n, nil
}

func (m *mockFile) Close() error {
	return nil // Mock implementation, does nothing
}

func TestCheckFileSignature_ValidAudioFile(t *testing.T) {
	// Create a context
	ctx := context.Background()

	audioFile := &mockFile{content: "\xFF\xFB" + strings.Repeat("\x00", helpers.SignatureLength-2)}

	// Pass context to the function
	err := helpers.CheckFileSignatureWithContext(ctx, audioFile)
	if err != nil {
		t.Errorf("Expected no error for valid audio file, got: %v", err)
	}
}

func TestCheckFileSignature_SmallFile(t *testing.T) {
	// Create a context
	ctx := context.Background()

	smallFile := &mockFile{content: strings.Repeat("\x00", helpers.SignatureLength-1)}

	// Pass context to the function
	err := helpers.CheckFileSignatureWithContext(ctx, smallFile)
	expectedError := "file size too small" // Update this based on the actual error message your function returns
	if err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error '%s' for small file, got: %v", expectedError, err)
	}
}

func TestCheckFileSignature_NonAudioFile(t *testing.T) {
	// Create a context
	ctx := context.Background()

	textFile := &mockFile{content: "Hello, world!" + strings.Repeat("\x00", helpers.SignatureLength-12)} // Adjusted for the new length check

	// Pass context to the function
	err := helpers.CheckFileSignatureWithContext(ctx, textFile)
	expectedError := "unknown file type" // Update this based on the actual error message your function returns
	if err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error '%s' for non-audio file, got: %v", expectedError, err)
	}
}

func TestCheckFileSignature_FilePointerResetAfterRead(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Simulate a file with a valid signature that is initially read, then fully processed
	// The mock file content includes a known valid file signature followed by arbitrary data
	validSignature := "\xFF\xFB" // Example valid signature for demonstration
	fileContent := validSignature + strings.Repeat("\x00", helpers.SignatureLength-2) + "Extra file content to simulate actual file data beyond the signature"
	file := &mockFile{content: fileContent}

	// Pass context and the mock file to the function
	err := helpers.CheckFileSignatureWithContext(ctx, file)
	if err != nil {
		t.Errorf("Expected no error after file signature read and reset, got: %v", err)
	}

	// Optionally, assert the file pointer is at the start after the check (if such an assertion is relevant to your logic)
	if file.offset != 0 {
		t.Errorf("Expected file pointer to be reset to the start after signature check, but it was at %d", file.offset)
	}
}
