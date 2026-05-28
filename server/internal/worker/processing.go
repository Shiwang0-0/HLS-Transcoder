package worker

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func (w *Worker) ProcessMessage(ctx context.Context, receiptHandle string, processFn func() error) error {
	heartbeatCtx, stopHeartbeat := context.WithCancel(ctx)
	defer stopHeartbeat()

	go func() {
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// if the ticker finished, increase the reset timeout
				_, err := w.SqsService.Client.ChangeMessageVisibility(
					heartbeatCtx, // use heartbeatCtx, not ctx
					&sqs.ChangeMessageVisibilityInput{
						QueueUrl:          &w.SqsService.QueueURL,
						ReceiptHandle:     &receiptHandle,
						VisibilityTimeout: *aws.Int32(60),
					},
				)
				if err != nil {
					log.Printf("Failed to extend visibility: %v", err)
				}
			case <-heartbeatCtx.Done():
				return
			}
		}
	}() // run this go routine
	return processFn() // run the generateTranscoding and uploadDirectory func that is passed to this function
	// when returned nil from there, this function ends, and Done channel gets invoked (stopHeartbeat() closes heartbeatCtx.Done() channel) and therefore the go routine too ends
}
