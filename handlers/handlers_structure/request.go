package handlers_structure

type SendData struct {
	BucketName string `json:"bucket"`
	FileName   string `json:"file_name"`
}
