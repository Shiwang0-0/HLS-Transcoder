package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/repository"
)

type JobService struct {
	JobRepository *repository.JobRepository
	SQSService    *SQSService
}

type JobStatusUpdate struct {
	Status string
	Stage  string
	Err    error
}

func NewJobService(jobRepository *repository.JobRepository, sqsService *SQSService) *JobService {
	return &JobService{
		JobRepository: jobRepository,
		SQSService:    sqsService,
	}
}

// send public job
func (s *JobService) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	job, err := s.JobRepository.GetJob(ctx, jobID)

	if err != nil {
		return nil, fmt.Errorf("error fetching job: %w", err)
	}
	return job, err
}

// WatchJob polls the DB until the job reaches a terminal state or ctx is cancelled,
// emitting an update only when status/stage actually changes.
func (s *JobService) WatchJob(ctx context.Context, jobID string) <-chan JobStatusUpdate {
	updates := make(chan JobStatusUpdate)

	go func() {
		defer close(updates)

		// Recover from any panic in this goroutine so a single bad request
		// can never take down the whole process — this is the critical fix.
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC recovered in WatchJob for jobID %s: %v", jobID, r)
				select {
				case updates <- JobStatusUpdate{Err: fmt.Errorf("internal error watching job")}:
				default:
				}
			}
		}()

		var lastStatus, lastStage string

		// check() returns true if the loop should stop (terminal state or error)
		check := func() bool {
			job, err := s.GetJob(ctx, jobID)
			if err != nil {
				select {
				case updates <- JobStatusUpdate{Err: err}:
				case <-ctx.Done():
				}
				return true
			}

			if job == nil {
				select {
				case updates <- JobStatusUpdate{Err: fmt.Errorf("job %s not found", jobID)}:
				case <-ctx.Done():
				}
				return true
			}

			if job.Status != lastStatus || job.Stage != lastStage {
				lastStatus, lastStage = job.Status, job.Stage
				select {
				case updates <- JobStatusUpdate{Status: job.Status, Stage: job.Stage}:
				case <-ctx.Done():
					return true
				}
			}

			return job.Status == "completed" || job.Status == "failed"
		}

		// immediate check on connect, don't make the client wait for the first tick
		if check() {
			return
		}

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if check() {
					return
				}
			}
		}
	}()

	return updates
}

// NotifyUploadToSQS updates job status in DB then pushes to SQS
func (s *JobService) UploadToSQS(ctx context.Context, job *models.Job) error {

	msg := &models.JobMessage{
		JobID:   job.JobID,
		VideoID: job.VideoID,
		Key:     job.Key,
	}

	if err := s.SQSService.PutInQueue(ctx, msg); err != nil {
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

func (s *JobService) QueueTranscodingJob(ctx context.Context, data models.JobCreationRequest) (*models.Job, error) {
	job, err := s.JobRepository.GetJobByVideoID(ctx, data.VideoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	if job.Status != "uploaded" {
		return nil, fmt.Errorf("upload not completed for videoID %s (status: %s)", data.VideoID, job.Status)
	}

	return job, nil
}

func (s *JobService) CreateJobShell(ctx context.Context, videoID, s3Key string) (string, error) {
	return s.JobRepository.CreateJobShell(ctx, videoID, s3Key)
}

func (s *JobService) MarkUploaded(ctx context.Context, videoID string) error {
	return s.JobRepository.UpdateJobStatusByVideoID(ctx, videoID, "uploaded", "s3_complete")
}

func (s *JobService) GetJobByVideoID(ctx context.Context, videoID string) (*models.Job, error) {
	return s.JobRepository.GetJobByVideoID(ctx, videoID)
}

// send public job
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
