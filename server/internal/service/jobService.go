package service

import (
	"context"
	"fmt"
	"log"
	"time"

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

	if err := s.SQSService.PutInQueue(ctx, job); err != nil {
		dbErr := s.JobRepository.UpdateJobStatus(ctx, job.JobID, "failed_to_queue", "sqs_error")
		if dbErr != nil {
			log.Printf("CRITICAL: Failed to update DB to 'failed_to_queue' for JobID %s: %v", job.JobID, dbErr)
		}
		return fmt.Errorf("sqs push failed: %w", err)
	}

	if err := s.JobRepository.UpdateJobStatus(ctx, job.JobID, "queued", "sqs"); err != nil {
		return fmt.Errorf("db update status failed: %w", err)
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

func (s *JobService) GetAllJobs(ctx context.Context) ([]*models.Job, error) {
	var jobs []*models.Job
	jobs, err := s.JobRepository.GetAllJobs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jobs: %w", err)
	}
	return jobs, nil
}

// retry go routine service

func (s *JobService) StartRetrySweeper(ctx context.Context) {
	// Wake up every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	log.Println("Starting background SQS retry sweeper...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down retry sweeper.")
			return
		case <-ticker.C:
			s.processFailedJobs(ctx)
		}
	}
}

func (s *JobService) processFailedJobs(ctx context.Context) {
	failedJobs, err := s.JobRepository.GetJobsByStatus(ctx, "failed_to_queue", 20) // fetch 20 of such jobs, limit to prevent memory spike
	if err != nil {
		log.Printf("Sweeper error fetching failed jobs: %v", err)
		return
	}

	if len(failedJobs) == 0 {
		return // Nothing to do
	}
	log.Printf("Sweeper found %d jobs failed_to_queue. Retrying...", len(failedJobs))

	for _, job := range failedJobs {
		err := s.UploadToSQS(ctx, job)
		if err != nil {
			log.Printf("Sweeper still failing for JobID %s: %v", job.JobID, err)
			continue
		}

		err = s.JobRepository.UpdateJobStatus(ctx, job.JobID, "queued", "recovered")
		if err != nil {
			log.Printf("Sweeper pushed to SQS but failed to update DB for JobID %s: %v", job.JobID, err)
		} else {
			log.Printf("Sweeper successfully recovered and queued JobID %s", job.JobID)
		}
	}
}
