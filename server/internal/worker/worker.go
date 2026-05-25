package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws/s3"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws/sqs"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/ffmpeg"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
)

type Worker struct {
	S3Service  *s3.Service
	SqsService *sqs.Service
}

func NewWorker(s3Service *s3.Service, sqsService *sqs.Service) *Worker {
	return &Worker{
		S3Service:  s3Service,
		SqsService: sqsService,
	}
}

func (w *Worker) Start(ctx context.Context) {
	for {
		pollCtx, cancel := context.WithTimeout(ctx, time.Second*60)

		/*
			you usually do NOT create ONE global timeout context for the entire application.
			Because: consumer should run forever
			optionally create short-lived contexts per request.
			so technically this ctx should be replaced with a pollCtx which has a timeout and it gets created for each iteration in infinte loop
		*/

		// poll sqs for messages
		result, err := w.SqsService.PollSQS(pollCtx)

		cancel() // cancle the short lived context
		if err != nil {
			continue
		}

		for _, msg := range result.Messages {
			var payload models.NotifyData

			err := json.Unmarshal(
				[]byte(*msg.Body),
				&payload,
			)

			if err != nil {
				log.Println(err)
				continue
			}

			objectKey := payload.Key

			// for every objectKey, get it from S3
			downloadCtx, cancel := context.WithTimeout(ctx, time.Second*60)
			localPath, err := w.S3Service.DownloadFile(downloadCtx, objectKey)

			cancel()

			if err != nil {
				log.Println(err)
				continue
			}

			err = ffmpeg.GenerateTranscoding(localPath)

			if err != nil {
				log.Print(err)
				continue
			}

			// delete after successful processing
			err = w.SqsService.DeleteMessage(ctx, *msg.ReceiptHandle)

			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}
