package dto

type PresignImageRequest struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
}

type PresignImageResponse struct {
	UploadURL string `json:"upload_url"`
	ObjectURL string `json:"object_url"`
	ObjectKey string `json:"object_key"`
}
