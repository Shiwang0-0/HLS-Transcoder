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
	query := `SELECT j.job_id, j.video_id, u.video_name, j.s3_key, j.status, j.stage, j.created_at FROM jobs j INNER JOIN uploads u ON j.video_id = u.video_id ORDER BY j.created_at DESC`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.Job

	for rows.Next() {
		job := &models.Job{}

		err := rows.Scan(&job.JobID, &job.VideoID, &job.VideoName, &job.Key, &job.Status, &job.Stage, &job.CreatedAt)
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

func (r *JobRepository) GetJobsByStatus(ctx context.Context, status string, limit int) ([]*models.Job, error) {
	query := `SELECT job_id, video_id, s3_key FROM jobs WHERE status = ? LIMIT ?`

	rows, err := r.DB.QueryContext(ctx, query, status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute select query: %w", err)
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		var j models.Job
		err := rows.Scan(&j.JobID, &j.VideoID, &j.Key)
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

func (r *JobRepository) CreateJob(ctx context.Context, jobID string, data models.JobCreationRequest) error {
	query := `INSERT INTO jobs (job_id, video_id, s3_key, status, stage) 
              VALUES (?, ?, ?, 'pending', 'creation')`

	_, err := r.DB.ExecContext(ctx, query, jobID, data.VideoID, data.Key)
	return err
}
