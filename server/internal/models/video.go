package models

type VideoMetadata struct {
	Name         string  `json:"name"`
	Size         int64   `json:"size"`
	Type         string  `json:"type"`
	Duration     float64 `json:"duration"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	LastModified int64   `json:"lastModified"`
}

type NotifyData struct {
	Key     string `json:"key"`
	JobID   string `json:"jobID"`
	VideoID string `json:"videoID"`
}

type PresignedURLResponse struct {
	URL     string `json:"url"`
	Key     string `json:"key"`
	VideoID string `json:"videoID"`
	JobID   string `json:"jobID"`
}

type JobStatus struct {
	JobID    string `json:"jobId"`
	Status   string `json:"status"` // uploading | queued | processing | completed | failed
	Stage    string `json:"stage"`  // s3_upload | ffmpeg | done
	Progress int    `json:"progress"`
	Key      string `json:"key"`
	Error    string `json:"error,omitempty"`
}
