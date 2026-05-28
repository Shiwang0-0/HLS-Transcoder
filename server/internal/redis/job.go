package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/redis/go-redis/v9"
)

type JobStore struct {
	rdb *redis.Client
}

func NewJobStore(rdb *redis.Client) *JobStore {
	return &JobStore{rdb: rdb}
}

func jobKey(jobID string) string {
	return fmt.Sprintf("job:%s", jobID)
}

// Create / Update job
func (j *JobStore) SetJob(ctx context.Context, job models.JobStatus) error {
	b, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return j.rdb.Set(ctx, jobKey(job.JobID), b, 24*time.Hour).Err()
}

// Get job
func (j *JobStore) GetJob(ctx context.Context, jobID string) (*models.JobStatus, error) {
	val, err := j.rdb.Get(ctx, jobKey(jobID)).Result()
	if err != nil {
		return nil, err
	}

	var job models.JobStatus
	if err := json.Unmarshal([]byte(val), &job); err != nil {
		return nil, err
	}

	return &job, nil
}

// Update partial status helper
func (j *JobStore) UpdateStatus(ctx context.Context, jobID string, status string, stage string, progress int) error {

	job, err := j.GetJob(ctx, jobID)
	if err != nil {
		return err
	}

	job.Status = status
	job.Stage = stage
	job.Progress = progress

	return j.SetJob(ctx, *job)
}
