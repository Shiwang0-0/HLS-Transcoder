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

func (r *JobRepository) GetAllJobs(ctx context.Context) ([]*models.Job, error) {
	query := `SELECT j.job_id, j.video_id, u.video_name, j.status, j.stage, j.created_at FROM jobs j INNER JOIN uploads u ON j.video_id = u.video_id ORDER BY j.created_at DESC`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.Job

	for rows.Next() {
		job := &models.Job{}

		err := rows.Scan(&job.JobID, &job.VideoID, &job.VideoName, &job.Status, &job.Stage, &job.CreatedAt)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

// used to push in sqs, so needs key therefore jobInternal
func (r *JobRepository) GetJobsByStatus(ctx context.Context, status string, limit int) ([]*models.JobInternal, error) {
	query := `SELECT job_id, s3_key, video_id FROM jobs WHERE status = ? LIMIT ?`

	rows, err := r.DB.QueryContext(ctx, query, status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute select query: %w", err)
	}
	defer rows.Close()

	var jobs []*models.JobInternal
	for rows.Next() {
		var j models.JobInternal
		err := rows.Scan(&j.JobID, &j.Key, &j.VideoID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job row: %w", err)
		}
		jobs = append(jobs, &j)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return jobs, nil
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

func (r *JobRepository) CreateJob(ctx context.Context, jobID string, data models.JobCreationRequest, key string) error {
	query := `INSERT INTO jobs (job_id, video_id, s3_key, status, stage) 
              VALUES (?, ?, ?, 'pending', 'creation')`

	_, err := r.DB.ExecContext(ctx, query, jobID, data.VideoID, key)
	return err
}

func (r *JobRepository) GetS3KeyByVideoID(ctx context.Context, videoID string) (string, error) {
	query := `SELECT s3_key FROM uploads WHERE video_id = ?`

	var s3Key string
	err := r.DB.QueryRowContext(ctx, query, videoID).Scan(&s3Key)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("upload session not found for videoID: %s", videoID)
		}
		return "", err
	}

	return s3Key, nil
}
