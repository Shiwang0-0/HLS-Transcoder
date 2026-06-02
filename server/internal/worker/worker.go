package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/ffmpeg"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/repository"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/service"
)

type Worker struct {
	S3Service     *service.S3Service
	SqsService    *service.SQSService
	JobRepository *repository.JobRepository
}

func NewWorker(s3Service *service.S3Service, sqsService *service.SQSService, jobRepository *repository.JobRepository) *Worker {
	return &Worker{
		S3Service:     s3Service,
		SqsService:    sqsService,
		JobRepository: jobRepository,
	}
}

func (w *Worker) Start(ctx context.Context) {
	fmt.Println("Worker starting...")
	for {
		pollCtx, cancelPoll := context.WithTimeout(ctx, time.Second*60)

		/*
			you usually do NOT create ONE global timeout context for the entire application.
			Because: consumer should run forever
			optionally create short-lived contexts per request.
			so technically this ctx should be replaced with a pollCtx which has a timeout and it gets created for each iteration in infinte loop
		*/

		// poll sqs for messages
		result, err := w.SqsService.PollSQS(pollCtx)
		cancelPoll() // cancle the short lived context
		if err != nil {
			log.Printf("SQS poll error: %v", err)
			continue
		}

		for _, msg := range result.Messages {
			var payload models.Job

			if err := json.Unmarshal([]byte(*msg.Body), &payload); err != nil {
				log.Println(err)
				continue
			}

			objectKey := payload.Key

			// update status to downloading
			w.JobRepository.UpdateJobStatus(ctx, payload.JobID, "downloading", "s3_download")

			downloadCtx, cancelDownload := context.WithTimeout(ctx, time.Second*120)
			localPath, err := w.S3Service.DownloadFile(downloadCtx, objectKey)
			cancelDownload()

			if err != nil {
				log.Printf("Download failed for %s: %v", objectKey, err)
				w.JobRepository.UpdateJobStatus(ctx, payload.JobID, "failed", "s3_download")
				continue
			}

			log.Printf("Downloaded to: %s", localPath)

			err = w.ProcessMessage(ctx, *msg.ReceiptHandle, func() error {
				w.JobRepository.UpdateJobStatus(ctx, payload.JobID, "transcoding", "ffmpeg")

				outputDir, err := ffmpeg.GenerateTranscoding(localPath, payload.VideoID)
				if err != nil {
					return err
				}

				log.Printf("Transcoding done. outputDir=%s videoID=%s", outputDir, payload.VideoID)

				w.JobRepository.UpdateJobStatus(ctx, payload.JobID, "uploading", "hls_upload")

				HLSKeyPrefix := "hls/" + payload.VideoID
				if err := w.S3Service.UploadDirectory(ctx, outputDir, HLSKeyPrefix); err != nil {
					return fmt.Errorf("upload failed: %w", err)
				}

				log.Printf("Successfully uploaded HLS for videoID: %s", payload.VideoID)
				return nil
			})

			if err != nil {
				log.Printf("processMessage failed: %v", err)
				w.JobRepository.UpdateJobStatus(ctx, payload.JobID, "failed", "processing")
				continue
			}

			if err := w.SqsService.DeleteMessage(ctx, *msg.ReceiptHandle); err != nil {
				log.Printf("Failed to delete SQS message: %v", err)
				continue
			}

			// mark as fully done
			w.JobRepository.UpdateJobStatus(ctx, payload.JobID, "completed", "done")
			log.Printf("Message deleted successfully")
		}
	}
}
