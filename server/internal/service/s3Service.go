package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/config"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

type S3Service struct {
	Client     *s3.Client
	BucketName string
}

func NewS3Service(client *s3.Client, bucketName string) *S3Service {
	return &S3Service{
		Client:     client,
		BucketName: bucketName,
	}
}

func (s *S3Service) GeneratePresignedPartURL(ctx context.Context, data models.PresignedPartURLRequest) (*models.PresignedURLResponse, error) {
	presignClient := s3.NewPresignClient(s.Client)

	req, err := presignClient.PresignUploadPart(
		ctx,
		&s3.UploadPartInput{
			Bucket: aws.String(s.BucketName),

			Key: aws.String(data.ObjectKey),

			UploadId: aws.String(data.UploadID),

			PartNumber: aws.Int32(
				int32(data.PartNumber),
			),
		},
		s3.WithPresignExpires(2*time.Minute),
	)

	if err != nil {
		return nil, err
	}
	return &models.PresignedURLResponse{
		URL: req.URL,
	}, nil
}

func (s *S3Service) InitMultipartUpload(ctx context.Context, data models.InitMultipartUploadRequest) (*models.UploadSession, error) {
	videoID := uuid.New().String()
	// objectKey is only of videoId
	ext := filepath.Ext(data.Name)
	objectKey := fmt.Sprintf("uploads/%s/source%s", videoID, ext)

	fmt.Println("MultiPart upload INIT objectKey: ", objectKey)

	// creates a multipart upload session and returns an uploadID, which will be used to upload the chunks
	result, err := s.Client.CreateMultipartUpload(
		ctx,
		&s3.CreateMultipartUploadInput{
			Bucket: aws.String(s.BucketName),
			Key:    aws.String(objectKey),

			ContentType: aws.String(data.Type),
		},
	)

	if err != nil {
		fmt.Println("S3 CreateMultipartUpload ERROR:", err)
		return nil, err
	}

	return &models.UploadSession{
		UploadID: aws.ToString(result.UploadId),
		Key:      objectKey,
		VideoID:  videoID,
		PartSize: config.DefaultPartSize,
		Status:   "uploading",
	}, nil
}

func (s *S3Service) AbortMultipartUpload(ctx context.Context, key, uploadID string) error {
	_, err := s.Client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(s.BucketName),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	})
	return err
}

func (s *S3Service) CompleteMultipartUpload(ctx context.Context, data models.CompleteMultipartUploadRequest) error {

	completedParts := make([]types.CompletedPart, 0)

	for _, part := range data.Parts {
		completedParts = append(completedParts,
			types.CompletedPart{
				ETag: aws.String(part.ETag),

				PartNumber: aws.Int32(
					part.PartNumber,
				),
			},
		)
	}

	_, err := s.Client.CompleteMultipartUpload(ctx,
		&s3.CompleteMultipartUploadInput{
			Bucket:   aws.String(s.BucketName),
			Key:      aws.String(data.Key),
			UploadId: aws.String(data.UploadID),
			MultipartUpload: &types.CompletedMultipartUpload{
				Parts: completedParts,
			},
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *S3Service) DownloadFile(ctx context.Context, objectKey string) (string, error) {

	// based on the size of the object that is on head, calculate the timeOutSeconds
	headResult, err := s.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &s.BucketName,
		Key:    aws.String(objectKey),
	})

	if err != nil {
		return "", err
	}

	fileSizeBytes := *headResult.ContentLength
	// assuming minimum 5 MB/s download speed, add 60s buffer
	timeoutSeconds := (fileSizeBytes / 1024 / 1024 / 5) + 60
	if timeoutSeconds < 60 {
		timeoutSeconds = 60
	}

	downloadCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	result, err := s.Client.GetObject(downloadCtx, &s3.GetObjectInput{
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

func (s *S3Service) UploadDirectory(ctx context.Context, localPath string, HLSKeyPrefix string) error {
	// recursive go into the folders
	return filepath.WalkDir(localPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err // propagate WalkDir errors
		}
		// only upload files, so skip directories
		if d.IsDir() {
			return nil
		}

		// based on localPath and currentPath get the relative path becuase S3 will include this relative path as the object key
		// converts media/uploads/abc123/720/segment000.ts to segment000.ts
		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		s3Key := HLSKeyPrefix + "/" + filepath.ToSlash(relPath)

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", path, err)
		}
		defer file.Close()

		contentType := "application/octet-stream"

		if strings.HasSuffix(path, ".m3u8") {
			contentType = "application/vnd.apple.mpegurl"
		}

		if strings.HasSuffix(path, ".ts") {
			contentType = "video/mp2t"
		}
		_, err = s.Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      &s.BucketName,
			Key:         &s3Key,
			Body:        file,
			ContentType: aws.String(contentType),
		})

		if err != nil {
			return fmt.Errorf("failed to upload %s: %w", s3Key, err)
		}

		fmt.Println("Uploaded:", s3Key)
		return nil
	})
}
