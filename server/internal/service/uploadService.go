package service

import (
	"context"
	"fmt"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/repository"
)

type UploadService struct {
	S3Service        *S3Service
	UploadRepository *repository.UploadRepository
}

func NewUploadService(uploadRepository *repository.UploadRepository, s3Service *S3Service) *UploadService {
	return &UploadService{
		UploadRepository: uploadRepository,
		S3Service:        s3Service,
	}
}

// GeneratePresignedPartURL delegates directly to S3
func (s *UploadService) GeneratePresignedPartURL(ctx context.Context, data models.PresignedPartURLRequest) (*models.PresignedURLResponse, error) {
	return s.S3Service.GeneratePresignedPartURL(ctx, data)
}

func (s *UploadService) InitMultipartUpload(ctx context.Context, data models.InitMultipartUploadRequest) (*models.UploadSession, error) {
	uploadSession, err := s.UploadRepository.CheckAlreadyUploaded(ctx, data.Fingerprint)
	if err != nil {
		return nil, fmt.Errorf("db check failed: %w", err)
	}

	if uploadSession != nil {
		return &models.UploadSession{
			Status:        uploadSession.Status,
			UploadID:      uploadSession.UploadID,
			Key:           uploadSession.Key,
			VideoID:       uploadSession.VideoID,
			PartSize:      uploadSession.PartSize,
			UploadedParts: uploadSession.UploadedParts,
		}, nil
	}

	// new upload
	uploadSession, err = s.S3Service.InitMultipartUpload(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("s3 init failed: %w", err)
	}

	if err := s.UploadRepository.CreateUploadSession(ctx, data, uploadSession); err != nil {
		s.S3Service.AbortMultipartUpload(ctx, uploadSession.Key, uploadSession.UploadID)
		return nil, fmt.Errorf("db create session failed: %w", err)
	}

	uploadSession.UploadedParts = []models.Part{}
	return uploadSession, nil
}

func (s *UploadService) CompleteMultipartUpload(ctx context.Context, data models.CompleteMultipartUploadRequest) error {
	if err := s.S3Service.CompleteMultipartUpload(ctx, data); err != nil {
		return fmt.Errorf("s3 complete failed: %w", err)
	}

	if err := s.UploadRepository.UpdateUploadSessionCompleted(ctx, data.UploadID, data.Parts); err != nil {
		fmt.Println("DB UpdateUploadSessionCompleted ERROR:", err) // don't fail — S3 already completed
	}

	return nil
}
