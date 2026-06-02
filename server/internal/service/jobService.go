package service

import (
	"context"
	"fmt"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/repository"
	"github.com/google/uuid"
)

type JobService struct {
	JobRepository *repository.JobRepository
	SQSService    *SQSService
}

func NewJobService(jobRepository *repository.JobRepository, sqsService *SQSService) *JobService {
	return &JobService{
		JobRepository: jobRepository,
		SQSService:    sqsService,
	}
}

func (s *JobService) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	job, err := s.JobRepository.GetJob(ctx, jobID)

	if err != nil {
		return nil, fmt.Errorf("error fetching job: %w", err)
	}
	return job, err
}

// NotifyUploadToSQS updates job status in DB then pushes to SQS
func (s *JobService) UploadToSQS(ctx context.Context, job *models.Job) error {
	if err := s.JobRepository.UpdateJobStatus(ctx, job.JobID, "queued", "sqs"); err != nil {
		return fmt.Errorf("db update status failed: %w", err)
	}

	if err := s.SQSService.PutInQueue(ctx, job); err != nil {
		return fmt.Errorf("sqs push failed: %w", err)
	}

	return nil
}

func (s *JobService) CreateTranscodingJob(ctx context.Context, data models.JobCreationRequest) (*models.Job, error) {

	jobID := uuid.New().String()

	err := s.JobRepository.CreateJob(ctx, jobID, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	job := &models.Job{
		JobID:   jobID,
		VideoID: data.VideoID,
		Key:     data.Key,
	}
	return job, nil
}
