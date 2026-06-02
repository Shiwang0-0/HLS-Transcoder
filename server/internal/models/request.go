package models

type PresignedPartURLRequest struct {
	ObjectKey  string `json:"objectKey"`
	UploadID   string `json:"uploadID"`
	PartNumber int64  `json:"partNumber"`
}

type JobCreationRequest struct {
	Key     string `json:"key"`
	VideoID string `json:"videoID"`
}

type InitMultipartUploadRequest struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Type        string `json:"type"`
	Fingerprint string `json:"fingerprint"`
}

type CompleteMultipartUploadRequest struct {
	UploadID string `json:"uploadID"`
	Key      string `json:"key"`
	VideoID  string `json:"videoID"`
	Parts    []Part `json:"parts"`
}
