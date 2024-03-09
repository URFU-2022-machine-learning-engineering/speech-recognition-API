package handlers_structure

type UploadError struct {
	Result string
}

type UploadSuccess struct {
	Result   string
	Filename string
}

type StatusSuccess struct {
	Status string `json:"status"`
}

type RecognitionSuccess struct {
	DetectedLang   string `json:"detected_language"`
	RecognizedText string `json:"recognized_text"`
}
