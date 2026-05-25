package sqs

import (
	"context"
	"encoding/json"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Service struct {
	Client   *sqs.Client
	QueueURL string
}

func NewService(sqsClient *sqs.Client, QueueURL string) *Service {
	return &Service{
		Client:   sqsClient,
		QueueURL: QueueURL,
	}
}

func (s *Service) PutInQueue(ctx context.Context, data models.NotifyData) error {

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

func (s *Service) PollSQS(ctx context.Context) (*sqs.ReceiveMessageOutput, error) {
	return s.Client.ReceiveMessage(
		ctx,
		&sqs.ReceiveMessageInput{
			QueueUrl:            &s.QueueURL,
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20,
		},
	)
}

func (s *Service) DeleteMessage(ctx context.Context, receiptHandle string) error {

	_, err := s.Client.DeleteMessage(
		ctx,
		&sqs.DeleteMessageInput{
			QueueUrl:      &s.QueueURL,
			ReceiptHandle: &receiptHandle,
		},
	)

	return err
}
