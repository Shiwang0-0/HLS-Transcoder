package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	mysqs "github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func PollSQS(ctx context.Context, sqsService *mysqs.Service) {
	for {

		pollCtx, cancel := context.WithTimeout(ctx, time.Second*60)
		result, err := sqsService.Client.ReceiveMessage(
			pollCtx,
			/*
				you usually do NOT create ONE global timeout context for the entire application.
				Because: consumer should run forever
				optionally create short-lived contexts per request.
				so technically this ctx should be replaced with a pollCtx which has a timeout and it gets created for each iteration in infinte loop
			*/
			&sqs.ReceiveMessageInput{
				QueueUrl:            &sqsService.QueueURL,
				MaxNumberOfMessages: 10,
				WaitTimeSeconds:     20,
			},
		)

		cancel() // cancle the short lived poll ctx

		if err != nil {
			log.Println(err)
			continue
		}

		for _, sqsMsg := range result.Messages {

			// process
			fmt.Println(*sqsMsg.Body)

			deleteCtx, deleteCancel := context.WithTimeout(
				ctx,
				10*time.Second,
			)

			// delete after success
			_, err := sqsService.Client.DeleteMessage(
				deleteCtx,
				&sqs.DeleteMessageInput{
					QueueUrl:      &sqsService.QueueURL,
					ReceiptHandle: sqsMsg.ReceiptHandle,
				},
			)

			deleteCancel()

			if err != nil {
				log.Println(err)
			}
		}
	}
}
