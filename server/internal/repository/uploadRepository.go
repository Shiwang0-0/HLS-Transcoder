package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
)

type UploadRepository struct {
	DB *sql.DB
}

func NewUploadRepository(db *sql.DB) *UploadRepository {
	return &UploadRepository{
		DB: db,
	}
}

func (r *UploadRepository) CreateUploadSession(ctx context.Context, data models.InitMultipartUploadRequest, session *models.UploadSession) error {
	query := `INSERT INTO uploads (fingerprint, video_id, upload_id, s3_key, part_size, uploaded_parts, status) 
	VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := r.DB.ExecContext(ctx, query,
		data.Fingerprint,
		session.VideoID,
		session.UploadID,
		session.Key,
		session.PartSize,
		"[]",
		session.Status,
	)
	return err
}

func (r *UploadRepository) CheckAlreadyUploaded(ctx context.Context, fingerprint string) (*models.UploadSession, error) {
	query := `SELECT upload_id, s3_key, video_id, status, part_size, uploaded_parts FROM uploads WHERE fingerprint = ? LIMIT 1`

	var session models.UploadSession

	var uploadedPartsRaw []byte

	err := r.DB.QueryRowContext(ctx, query, fingerprint).Scan(
		&session.UploadID,
		&session.Key,
		&session.VideoID,
		&session.Status,
		&session.PartSize,
		&uploadedPartsRaw,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if len(uploadedPartsRaw) > 0 {
		json.Unmarshal(uploadedPartsRaw, &session.UploadedParts)
	} else {
		session.UploadedParts = []models.Part{}
	}

	return &session, nil
}

func (r *UploadRepository) UpdateUploadSessionCompleted(ctx context.Context, uploadID string, parts []models.Part) error {
	partsJSON, err := json.Marshal(parts)
	if err != nil {
		return err
	}

	query := `UPDATE uploads SET status = 'completed', uploaded_parts = ? WHERE upload_id = ?`

	_, err = r.DB.ExecContext(ctx, query, string(partsJSON), uploadID)
	return err
}
