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
	Key string `json:"key"`
}

type PresignedURLResponse struct {
	URL string `json:"url"`
	Key string `json:"key"`
}
