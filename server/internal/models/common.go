package models

import "time"

type Part struct {
	ETag       string `json:"ETag"`
	PartNumber int32  `json:"PartNumber"`
}

type UploadSession struct {
	ID            int64   `json:"id"` // session row id in db
	VideoID       string  `json:"videoID"`
	Status        string  `json:"status"`
	PartSize      float64 `json:"partSize"`
	UploadedParts []Part  `json:"uploadedParts"`
}

type UploadSessionInternal struct {
	ID            int64
	UploadID      string
	Key           string
	VideoID       string
	Status        string
	PartSize      float64
	UploadedParts []Part
}

type Job struct {
	JobID     string    `json:"jobID"`
	VideoID   string    `json:"videoID"`
	VideoName string    `json:"videoName"`
	Status    string    `json:"status"` // queued | processing | completed | failed
	Stage     string    `json:"stage"`  // sqs | ffmpeg | done
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type JobInternal struct {
	Key       string    `json:"key"`
	JobID     string    `json:"jobID"`
	VideoID   string    `json:"videoID"`
	VideoName string    `json:"videoName"`
	Status    string    `json:"status"` // queued | processing | completed | failed
	Stage     string    `json:"stage"`  // sqs | ffmpeg | done
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

func (s *UploadSessionInternal) ToPublic() UploadSession {
	return UploadSession{
		ID:            s.ID,
		VideoID:       s.VideoID,
		Status:        s.Status,
		PartSize:      s.PartSize,
		UploadedParts: s.UploadedParts,
	}
}

func (s *JobInternal) ToPublic() Job {
	return Job{
		JobID:     s.JobID,
		VideoID:   s.VideoID,
		VideoName: s.VideoName,
		Status:    s.Status,
		Stage:     s.Stage,
		Error:     s.Error,
		CreatedAt: s.CreatedAt,
	}
}
