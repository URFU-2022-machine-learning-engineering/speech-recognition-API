package tests

// Define a mock implementation of multipart.File for testing purposes
type mockFile struct {
	content string
	offset  int64
}
