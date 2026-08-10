package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/google/uuid"
)

type JobRepository struct {
	DB *sql.DB
}

type UploadStatus struct {
	S3Key  string
	Status string
}

func NewJobRepository(db *sql.DB) *JobRepository {
	return &JobRepository{
		DB: db,
	}
}

// CreateJobShell creates a job row the moment an upload session starts,
// before any parts or completion has happened. Key is already known at
// this point (S3Service.InitMultipartUpload generates it deterministically).
func (r *JobRepository) CreateJobShell(ctx context.Context, videoID, s3Key string) (string, error) {
	jobID := uuid.New().String()

	query := `INSERT INTO jobs (job_id, video_id, s3_key, status, stage)
	          VALUES (?, ?, ?, 'pending_upload', 'init')`

	_, err := r.DB.ExecContext(ctx, query, jobID, videoID, s3Key)
	if err != nil {
		return "", err
	}
	return jobID, nil
}

func (r *JobRepository) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	query := `SELECT job_id, video_id, status, stage FROM jobs WHERE job_id = ?`

	var job models.Job

	err := r.DB.QueryRowContext(ctx, query, jobID).
		Scan(&job.JobID, &job.VideoID, &job.Status, &job.Stage)

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

// used to push in sqs, so needs key therefore Job
func (r *JobRepository) GetJobsByStatus(ctx context.Context, status string, limit int) ([]*models.Job, error) {
	query := `SELECT job_id, s3_key, video_id FROM jobs WHERE status = ? LIMIT ?`

	rows, err := r.DB.QueryContext(ctx, query, status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute select query: %w", err)
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		var j models.Job
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

func (r *JobRepository) GetJobByVideoID(ctx context.Context, videoID string) (*models.Job, error) {
	query := `SELECT job_id, video_id, s3_key, status, stage FROM jobs WHERE video_id = ?`

	var job models.Job
	err := r.DB.QueryRowContext(ctx, query, videoID).Scan(
		&job.JobID, &job.VideoID, &job.Key, &job.Status, &job.Stage,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no job found for videoID %s", videoID)
		}
		return nil, err
	}
	return &job, nil
}

func (r *JobRepository) UpdateJobStatusByVideoID(ctx context.Context, videoID, status, stage string) error {
	query := `UPDATE jobs SET status = ?, stage = ? WHERE video_id = ?`
	_, err := r.DB.ExecContext(ctx, query, status, stage, videoID)
	return err
}
