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
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

type S3Service struct {
	Client          *s3.Client
	BucketName      string
	TransferManager *transfermanager.Client
}

func NewS3Service(client *s3.Client, bucketName string) *S3Service {
	return &S3Service{
		Client:     client,
		BucketName: bucketName,
		TransferManager: transfermanager.New(client, func(o *transfermanager.Options) {
			o.PartSizeBytes = 5 * 1024 * 1024
			o.Concurrency = 5 // 5 concurrent network pipes
		}),
	}
}

func (s *S3Service) GeneratePresignedPartURL(ctx context.Context, data models.PresignedPartURLRequest, objectKey, uploadId string) (*models.PresignedURLResponse, error) {
	presignClient := s3.NewPresignClient(s.Client)

	req, err := presignClient.PresignUploadPart(
		ctx,
		&s3.UploadPartInput{
			Bucket: aws.String(s.BucketName),

			Key: aws.String(objectKey),

			UploadId: aws.String(uploadId),

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

func (s *S3Service) CompleteMultipartUpload(ctx context.Context, data models.CompleteMultipartUploadRequest, key, uploadID string) error {

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
			Key:      aws.String(key),
			UploadId: aws.String(uploadID),
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

// download large files on concurrent network pipes
func (s *S3Service) DownloadFile(ctx context.Context, objectKey string) (string, error) {
	localPath := filepath.Join("media", objectKey)
	parentDir := filepath.Dir(localPath)

	// Ensure the destination directory exists
	if err := os.MkdirAll(parentDir, os.ModePerm); err != nil {
		log.Printf("Couldn't create directory %v, error: %v\n", parentDir, err)
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(localPath)
	if err != nil {
		log.Printf("Couldn't create file %v. Here's why: %v\n", localPath, err)
		return "", fmt.Errorf("failed to create target file: %w", err)
	}

	defer file.Close()

	log.Printf("Starting transfermanager stream download for: %s\n", objectKey)
	result, err := s.TransferManager.GetObject(ctx, &transfermanager.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		_ = os.Remove(localPath)

		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			log.Printf("Can't get object %s from bucket %s. No such key exists.\n", objectKey, s.BucketName)
			return "", noKey
		}

		log.Printf("Couldn't initialize stream for %v:%v. Error: %v\n", s.BucketName, objectKey, err)
		return "", fmt.Errorf("transfermanager GetObject failed: %w", err)
	}

	// Use io.Copy to pump the data out of the managed network pool into the file
	_, err = io.Copy(file, result.Body)
	if err != nil {
		_ = os.Remove(localPath) // Clean up if network drops mid-stream
		log.Printf("Failed to write stream to file: %v\n", err)
		return "", fmt.Errorf("file write stream failed: %w", err)
	}

	log.Printf("Successfully downloaded %s to %s\n", objectKey, localPath)
	return localPath, nil
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

func (s *S3Service) ListPartsFromS3(ctx context.Context, uploadID, key string) ([]models.Part, error) {
	var allParts []models.Part
	var partNumberMarker *string

	for {
		input := &s3.ListPartsInput{
			Bucket:           aws.String(s.BucketName),
			Key:              aws.String(key),
			UploadId:         aws.String(uploadID),
			PartNumberMarker: partNumberMarker,
		}

		output, err := s.Client.ListParts(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("s3 ListParts: %w", err)
		}

		for _, p := range output.Parts {
			allParts = append(allParts, models.Part{
				PartNumber: *p.PartNumber,
				ETag:       aws.ToString(p.ETag),
			})
		}

		if output.IsTruncated == nil || !*output.IsTruncated {
			break
		}
		partNumberMarker = output.NextPartNumberMarker
	}

	return allParts, nil
}
