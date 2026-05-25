package worker

import (
	"context"
	"time"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws/s3"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws/sqs"
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
			objectKey := *msg.Body

			// for every objectKey, get it from S3
			downloadCtx, cancel := context.WithTimeout(ctx, time.Second*60)
			w.S3Service.DownloadFile(downloadCtx, objectKey)

			cancel()
		}
	}
}
