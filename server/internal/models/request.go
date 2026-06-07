package models

type PresignedPartURLRequest struct {
	SessionID  int64 `json:"sessionID"`
	PartNumber int64 `json:"PartNumber"`
}

type JobCreationRequest struct {
	SessionID int64  `json:"sessionID"`
	VideoID   string `json:"videoID"`
}

type InitMultipartUploadRequest struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Type        string `json:"type"`
	Fingerprint string `json:"fingerprint"`
}

type CompleteMultipartUploadRequest struct {
	SessionID int64  `json:"sessionID"`
	VideoID   string `json:"videoID"`
	Parts     []Part `json:"parts"`
}
