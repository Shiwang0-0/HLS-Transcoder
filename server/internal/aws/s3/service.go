package s3

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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

func (s *Service) DownloadFile(ctx context.Context, objectKey string) error {
	result, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			log.Printf("Can't get object %s from bucket %s. No such key exists.\n", objectKey, s.BucketName)
			err = noKey
		} else {
			log.Printf("Couldn't get object %v:%v. Here's why: %v\n", s.BucketName, objectKey, err)
		}
		return err
	}
	defer result.Body.Close()

	dir := "./media"

	// create directory if not exists
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Printf("Couldn't create directory %v. Here's why: %v\n", dir, err)
		return err
	}

	fileName := filepath.Join(dir, objectKey)
	// save file
	file, err := os.Create(fileName)
	if err != nil {
		log.Printf("Couldn't create file %v. Here's why: %v\n", fileName, err)
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, result.Body)
	return err
}
