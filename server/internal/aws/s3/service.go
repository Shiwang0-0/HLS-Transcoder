package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
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

func (s *Service) GeneratePresignedURL(ctx context.Context, metaData models.VideoMetadata) (*models.PresignedURLResponse, error) {
	presignClient := s3.NewPresignClient(s.Client)

	// for every video to be identified as uniue, add uuid in the objectKey
	videoID := uuid.NewString()

	objectKey := fmt.Sprintf(
		"input/%s/%s",
		videoID,
		metaData.Name,
	)

	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.BucketName),
		Key:         aws.String(objectKey),
		ContentType: aws.String(metaData.Type),
	}, s3.WithPresignExpires(2*time.Minute))

	if err != nil {
		return nil, err
	}
	return &models.PresignedURLResponse{
		URL: req.URL,
		Key: objectKey,
	}, nil
}

func (s *Service) DownloadFile(ctx context.Context, objectKey string) (string, error) {
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
		return "", err
	}
	defer result.Body.Close()

	localPath := filepath.Join("media", objectKey)

	parentDir := filepath.Dir(localPath)

	err = os.MkdirAll(parentDir, os.ModePerm)
	if err != nil {
		log.Printf("couldn't create directory %v, error: %v\n", parentDir, err)
		return "", err
	}

	// save file
	file, err := os.Create(localPath)
	if err != nil {
		log.Printf("Couldn't create file %v. Here's why: %v\n", localPath, err)
		return "", err
	}
	defer file.Close()
	_, err = io.Copy(file, result.Body)
	return localPath, err
}

func (s *Service) UploadDirectory(ctx context.Context, localPath string, HLSKeyPrefix string) error {
	// recursive go into the folders
	return filepath.WalkDir(localPath, func(path string, d os.DirEntry, err error) error {
		// only upload files, so skip directories
		if d.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		relPath, err := filepath.Rel(localPath, path) // based on localPath and currentPath get the relative path becuase S3 will include this relative path as the object key
		// converts media/uploads/abc123/720/segment000.ts to segment000.ts

		s3Key := HLSKeyPrefix + "/" + filepath.ToSlash(relPath)

		_, err = s.Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &s.BucketName,
			Key:    &s3Key,
			Body:   file,
		})

		if err != nil {
			return err
		}

		fmt.Println("Uploaded:", s3Key)

		return nil
	})
}
