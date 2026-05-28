package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws/s3"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws/sqs"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/ffmpeg"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/redis"
)

type Worker struct {
	S3Service  *s3.Service
	SqsService *sqs.Service
	JobStore   *redis.JobStore
}

func NewWorker(s3Service *s3.Service, sqsService *sqs.Service, JobStore *redis.JobStore) *Worker {
	return &Worker{
		S3Service:  s3Service,
		SqsService: sqsService,
		JobStore:   JobStore,
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

		for _, msg := range result.Messages { // for every video
			var payload models.NotifyData

			if err := json.Unmarshal([]byte(*msg.Body), &payload); err != nil {
				log.Println(err)
				continue
			}

			objectKey := payload.Key

			// for every objectKey, get it from S3
			downloadCtx, cancelDownload := context.WithTimeout(ctx, time.Second*120)
			localPath, err := w.S3Service.DownloadFile(downloadCtx, objectKey)
			cancelDownload() // cancel immediately after download

			if err != nil {
				log.Printf("Download failed for %s: %v", objectKey, err)
				continue
			}

			log.Printf("Downloaded to: %s", localPath)

			// this process Message works on heartbeat,
			// so every 20 seconds the sqs VisibilityTimeout increases because of the long processing task
			// root ctx passed here, heartbeat manages SQS timeout internally

			w.JobStore.UpdateStatus(ctx, payload.JobID, "processing", "ffmpeg", 10)
			err = w.ProcessMessage(ctx, *msg.ReceiptHandle, func() error {
				// Both inside processMessage — heartbeat covers everything
				outputDir, err := ffmpeg.GenerateTranscoding(localPath, payload.VideoID)
				if err != nil {
					return err
				}

				log.Printf("Transcoding done. outputDir=%s videoID=%s", outputDir, payload.VideoID)

				// because the uploading part is also a long one, use heartbeat here also
				HLSKeyPrefix := "hls/" + payload.VideoID
				if err := w.S3Service.UploadDirectory(ctx, outputDir, HLSKeyPrefix); err != nil {
					return fmt.Errorf("upload failed: %w", err)
				}

				log.Printf("Successfully uploaded HLS for videoID: %s", payload.VideoID)
				return nil
			})

			if err != nil {
				log.Printf("processMessage failed: %v", err)
				continue
			}

			// Delete ONLY after both transcode + upload succeed
			if err := w.SqsService.DeleteMessage(ctx, *msg.ReceiptHandle); err != nil {
				log.Printf("Failed to delete SQS message: %v", err)
				continue
			}

			log.Printf("Message deleted successfully")

			job := models.JobStatus{
				JobID:    payload.JobID,
				Status:   "completed",
				Stage:    "done",
				Progress: 100,
			}
			w.JobStore.SetJob(ctx, job)
		}
	}
}
