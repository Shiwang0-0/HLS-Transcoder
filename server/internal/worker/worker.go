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
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
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

		/*
			you usually do NOT create ONE global timeout context for the entire application.
			Because: consumer should run forever
			optionally create short-lived contexts per request.
			so technically this ctx should be replaced with a pollCtx which has a timeout and it gets created for each iteration in infinte loop
		*/

		// poll sqs for messages

		pollCtx, cancelPoll := context.WithTimeout(ctx, time.Second*25) // SQS long polling caps at 20 seconds
		result, err := w.SqsService.PollSQS(pollCtx)
		cancelPoll()
		if err != nil {
			log.Printf("SQS poll error: %v", err)
			continue
		}

		for _, msg := range result.Messages {
			w.handleMessage(ctx, msg)
		}
	}
}
func (w *Worker) handleMessage(ctx context.Context, msg types.Message) {
	var payload models.JobMessage
	if err := json.Unmarshal([]byte(*msg.Body), &payload); err != nil {
		log.Println(err)
		return
	}

	// heartbeat scoped to THIS message's receipt handle
	stop := w.heartbeatVisibility(ctx, *msg.ReceiptHandle, 10*time.Minute, 30*time.Minute)
	defer stop() // fires when handleMessage returns — covers every exit path below

	objectKey := payload.Key
	w.JobRepository.UpdateJobStatus(ctx, payload.JobID, "downloading", "s3_download")

	downloadCtx, cancelDownload := context.WithTimeout(ctx, time.Second*120)
	localPath, err := w.S3Service.DownloadFile(downloadCtx, objectKey)
	cancelDownload()
	if err != nil {
		log.Printf("Download failed for %s: %v", objectKey, err)
		w.JobRepository.UpdateJobStatus(ctx, payload.JobID, "failed", "s3_download")
		return
	}

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
		return
	}

	if err := w.SqsService.DeleteMessage(ctx, *msg.ReceiptHandle); err != nil {
		log.Printf("Failed to delete SQS message: %v", err)
		return
	}

	w.JobRepository.UpdateJobStatus(ctx, payload.JobID, "completed", "done")
	log.Printf("Message deleted successfully")
}

func (w *Worker) heartbeatVisibility(ctx context.Context, receiptHandle string, interval, extendBy time.Duration) (stop func()) {
	hbCtx, cancle := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-hbCtx.Done():
				return

			case <-ticker.C:
				if err := w.SqsService.ChangeMessageVisibility(hbCtx, receiptHandle, extendBy); err != nil {
					log.Printf("failed to extend visibility: %v", err)
				}
			}
		}
	}()
	return cancle
}
