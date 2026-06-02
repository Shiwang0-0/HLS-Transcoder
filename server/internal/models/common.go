package models

type Part struct {
	ETag       string `json:"ETag"`
	PartNumber int32  `json:"PartNumber"`
}

type UploadSession struct {
	UploadID      string  `json:"uploadID"`
	Key           string  `json:"key"`
	VideoID       string  `json:"videoID"`
	Status        string  `json:"status"`
	PartSize      float64 `json:"partSize"`
	UploadedParts []Part  `json:"uploadedParts"`
}

type Job struct {
	JobID   string `json:"jobID"`
	VideoID string `json:"videoID"`
	Key     string `json:"key"`
	Status  string `json:"status"` // queued | processing | completed | failed
	Stage   string `json:"stage"`  // sqs | ffmpeg | done
	Error   string `json:"error,omitempty"`
}
