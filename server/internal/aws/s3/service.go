package s3

import (
	"context"
	"time"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Service struct {
	Client     *s3.Client
	BucketName string
}

func NewService(s3Client *s3.Client, bucketName string) *Service {
	return &Service{
		Client:     s3Client,
		BucketName: bucketName,
	}
}

func (s *Service) GeneratePresignedURL(ctx context.Context, metaData models.VideoMetadata) (string, error) {
	presignClient := s3.NewPresignClient(s.Client)

	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.BucketName),
		Key:         aws.String(metaData.Name),
		ContentType: aws.String(metaData.Type),
	}, s3.WithPresignExpires(2*time.Minute))

	if err != nil {
		return "", err
	}
	return req.URL, nil
}
