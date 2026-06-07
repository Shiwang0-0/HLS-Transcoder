package models

type PresignedURLResponse struct {
	URL     string `json:"url"`
	VideoID string `json:"videoID"`
	JobID   string `json:"jobID"`
}

type JobCreationResponse struct {
	JobID   string `json:"jobID"`
	VideoID string `json:"videoID"`
}

type InitMultipartUploadResponse struct {
	Status        string  `json:"status"`
	VideoID       string  `json:"videoID"`
	JobID         string  `json:"jobID"`
	PartSize      float64 `json:"partSize"`
	UploadedParts []Part  `json:"uploadedParts"`
}
