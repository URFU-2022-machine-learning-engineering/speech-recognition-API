package handlers_structure

type UploadError struct {
	Result string
}

type UploadSuccess struct {
	Result   string
	Filename string
}

type StatusSuccess struct {
	Status string
}
