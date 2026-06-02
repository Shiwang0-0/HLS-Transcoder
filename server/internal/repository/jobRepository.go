package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
)

type JobRepository struct {
	DB *sql.DB
}

func NewJobRepository(db *sql.DB) *JobRepository {
	return &JobRepository{
		DB: db,
	}
}

func (r *JobRepository) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	query := `SELECT video_id, status, stage FROM jobs WHERE job_id = ?`

	var job models.Job

	err := r.DB.QueryRowContext(ctx, query, jobID).
		Scan(&job.VideoID, &job.Status, &job.Stage)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	fmt.Printf("job: %v", job)

	return &job, nil
}

func (r *JobRepository) UpdateJobStatus(ctx context.Context, jobID, status, stage string) error {

	query := `UPDATE jobs SET status = ?, stage = ? WHERE job_id = ?`
	_, err := r.DB.ExecContext(ctx, query, status, stage, jobID)
	if err != nil {
		log.Printf("Failed to update job status for %s: %v", jobID, err)
		return err
	}
	return nil
}

func (r *JobRepository) CreateJob(ctx context.Context, jobID string, data models.JobCreationRequest) error {
	query := `INSERT INTO jobs (job_id, video_id, s3_key, status, stage) 
              VALUES (?, ?, ?, 'queued', 'waiting')`

	_, err := r.DB.ExecContext(ctx, query, jobID, data.VideoID, data.Key)
	return err
}
