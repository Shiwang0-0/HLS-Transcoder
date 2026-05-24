package sqs

import (
	"context"

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
	_, err := s.Client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &s.QueueURL,
		MessageBody: aws.String(string(data.Key)),
	})

	if err != nil {
		return err
	}

	return nil
}
