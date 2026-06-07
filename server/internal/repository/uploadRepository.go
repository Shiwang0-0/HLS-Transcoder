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

func (r *UploadRepository) CreateUploadSession(ctx context.Context, data models.InitMultipartUploadRequest, session *models.UploadSessionInternal) error {
	query := `INSERT INTO uploads (fingerprint, video_name, video_id, upload_id, s3_key, part_size, uploaded_parts, status) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.DB.ExecContext(ctx, query,
		data.Fingerprint,
		data.Name,
		session.VideoID,
		session.UploadID,
		session.Key,
		session.PartSize,
		"[]",
		session.Status,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	session.ID = id // session ID on session creation in db
	return nil
}

func (r *UploadRepository) CheckAlreadyUploaded(ctx context.Context, fingerprint string) (*models.UploadSessionInternal, error) {
	query := `SELECT id, upload_id, s3_key, video_id, status, part_size, uploaded_parts FROM uploads WHERE fingerprint = ? LIMIT 1`

	var session models.UploadSessionInternal

	var uploadedPartsRaw []byte

	err := r.DB.QueryRowContext(ctx, query, fingerprint).Scan(
		&session.ID,
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

func (r *UploadRepository) GetSessionByID(ctx context.Context, sessionID int64) (*models.UploadSessionInternal, error) {
	query := `SELECT upload_id, s3_key, uploaded_parts FROM uploads WHERE id = ?`

	var session models.UploadSessionInternal

	var uploadedPartsRaw []byte

	err := r.DB.QueryRowContext(ctx, query, sessionID).Scan(
		&session.UploadID,
		&session.Key,
		&uploadedPartsRaw,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if len(uploadedPartsRaw) > 0 {
		if err := json.Unmarshal(uploadedPartsRaw, &session.UploadedParts); err != nil {
			return nil, err
		}
	} else {
		session.UploadedParts = []models.Part{}
	}

	return &session, nil
}

func (r *UploadRepository) VerifyAndPersistParts(ctx context.Context, allVerifiedParts []models.Part, sessionID int64) error {

	partsJSON, err := json.Marshal(allVerifiedParts)
	if err != nil {
		return err
	}

	query := `UPDATE uploads SET uploaded_parts = ? where id = ?`

	_, err = r.DB.ExecContext(ctx, query, string(partsJSON), sessionID)
	return err
}

func (r *UploadRepository) UpdateUploadSessionCompleted(ctx context.Context, sessionID int64, parts []models.Part) error {
	partsJSON, err := json.Marshal(parts)
	if err != nil {
		return err
	}

	query := `UPDATE uploads SET status = 'completed', uploaded_parts = ? WHERE id = ?`

	_, err = r.DB.ExecContext(ctx, query, string(partsJSON), sessionID)
	return err
}
