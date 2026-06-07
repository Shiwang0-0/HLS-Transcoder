package service

import (
	"context"
	"encoding/json"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSService struct {
	Client   *sqs.Client
	QueueURL string
}

func NewSQSService(client *sqs.Client, queueURL string) *SQSService {
	return &SQSService{
		Client:   client,
		QueueURL: queueURL,
	}
}

func (s *SQSService) PutInQueue(ctx context.Context, data *models.JobInternal) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = s.Client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &s.QueueURL,
		MessageBody: aws.String(string(body)),
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *SQSService) PollSQS(ctx context.Context) (*sqs.ReceiveMessageOutput, error) {
	return s.Client.ReceiveMessage(
		ctx,
		&sqs.ReceiveMessageInput{
			QueueUrl:            &s.QueueURL,
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20,
		},
	)
}

func (s *SQSService) DeleteMessage(ctx context.Context, receiptHandle string) error {

	_, err := s.Client.DeleteMessage(
		ctx,
		&sqs.DeleteMessageInput{
			QueueUrl:      &s.QueueURL,
			ReceiptHandle: &receiptHandle,
		},
	)

	return err
}
