package models

import "time"

type Part struct {
	ETag       string `json:"ETag"`
	PartNumber int32  `json:"PartNumber"`
}

type UploadSession struct {
	ID            int64   `json:"id"`
	VideoID       string  `json:"videoID"`
	JobID         string  `json:"jobID"`
	Key           string  `json:"-"`      // internal only
	UploadID      string  `json:"-"`      // internal only
	Status        string  `json:"status"` // uploading | completed
	PartSize      float64 `json:"partSize"`
	UploadedParts []Part  `json:"uploadedParts"`
	IsNewSession  bool    `json:"-"` // internal signal
}

type Job struct {
	JobID     string    `json:"jobID"`
	VideoID   string    `json:"videoID"`
	VideoName string    `json:"videoName"`
	Key       string    `json:"-"`      // internal only
	Status    string    `json:"status"` // pending_upload | uploaded | queued | processing | completed | failed
	Stage     string    `json:"stage"`  // init | s3_complete | sqs | download | ffmpeg | done
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// JobMessage is what actually goes into the SQS message body
// separate from Job's HTTP-response shape, since the worker needs Key.
type JobMessage struct {
	JobID   string `json:"jobID"`
	VideoID string `json:"videoID"`
	Key     string `json:"key"`
}
