package service

import (
	"context"
	"fmt"
	"sort"

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

	session, err := s.UploadRepository.GetSessionByID(ctx, data.SessionID)

	if err != nil {
		return nil, fmt.Errorf("db check failed: %w", err)
	}

	return s.S3Service.GeneratePresignedPartURL(ctx, data, session.Key, session.UploadID)
}

func (s *UploadService) InitMultipartUpload(ctx context.Context, data models.InitMultipartUploadRequest) (*models.UploadSessionInternal, error) {
	uploadSession, err := s.UploadRepository.CheckAlreadyUploaded(ctx, data.Fingerprint)
	if err != nil {
		return nil, fmt.Errorf("db check failed: %w", err)
	}

	if uploadSession != nil {
		return uploadSession, nil // already UploadSessionInternal, return as-is
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

func (s *UploadService) VerifyAndPersistParts(ctx context.Context, sessionID int64, parts []models.Part) ([]models.Part, []models.Part, error) {

	// get session from the DB
	session, err := s.UploadRepository.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("DB error: %w", err)
	}
	if session == nil {
		return nil, nil, fmt.Errorf("upload session not found: %d", sessionID)
	}

	if len(parts) == 0 { // no parts were uploaded by user, return the already existing parts in DB
		return session.UploadedParts, []models.Part{}, nil
	}

	var existingParts []models.Part = session.UploadedParts

	s3Parts, err := s.S3Service.ListPartsFromS3(ctx, session.UploadID, session.Key)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to List parts from S3: %w", err)
	}

	s3ETagMap := make(map[int32]string, len(s3Parts))
	for _, p := range s3Parts {
		s3ETagMap[p.PartNumber] = p.ETag
	}

	verifiedParts := make([]models.Part, 0)
	missingParts := make([]models.Part, 0)

	// iterating over all parts uploded by client
	for _, claim := range parts {
		actualETag, exists := s3ETagMap[claim.PartNumber]
		if exists && actualETag == claim.ETag {
			// uploaded by client, found on S3
			verifiedParts = append(verifiedParts, models.Part{
				PartNumber: claim.PartNumber,
				ETag:       claim.ETag,
			})
		} else {
			// ulpoaded by client, doesnt exist on S3
			missingParts = append(missingParts, models.Part{
				PartNumber: claim.PartNumber,
				ETag:       claim.ETag,
			})
		}
	}

	// the exisiting parts (the one in DB, these are verified) and verified parts (which are just uploaded yet but already exist on S3)
	// verifiedParts may include parts that are already in the DB but were recently uploaded again (so remove duplicates)
	existingSet := make(map[int32]bool, len(existingParts)) // get all the DB stored ones
	for _, p := range existingParts {
		existingSet[p.PartNumber] = true
	}
	allVerifiedParts := make([]models.Part, len(existingParts))
	copy(allVerifiedParts, existingParts)

	for _, p := range verifiedParts {
		if !existingSet[p.PartNumber] {
			allVerifiedParts = append(allVerifiedParts, p) // get all the verified ones that are not stored in DB yet but only on S3
		}
	}

	sort.Slice(allVerifiedParts, func(i, j int) bool {
		return allVerifiedParts[i].PartNumber < allVerifiedParts[j].PartNumber
	})

	// persist the parts that are uploaded to db
	if err := s.UploadRepository.VerifyAndPersistParts(ctx, allVerifiedParts, sessionID); err != nil {
		fmt.Println("DB VerifyAndPersistParts ERROR:", err)
		return nil, nil, err
	}

	return allVerifiedParts, missingParts, nil
}

func (s *UploadService) CompleteMultipartUpload(ctx context.Context, data models.CompleteMultipartUploadRequest) error {

	session, err := s.UploadRepository.GetSessionByID(ctx, data.SessionID)

	if err != nil {
		return fmt.Errorf("db check failed: %w", err)
	}

	if err := s.S3Service.CompleteMultipartUpload(ctx, data, session.Key, session.UploadID); err != nil {
		return fmt.Errorf("s3 complete failed: %w", err)
	}

	if err := s.UploadRepository.UpdateUploadSessionCompleted(ctx, data.SessionID, data.Parts); err != nil {
		fmt.Println("DB UpdateUploadSessionCompleted ERROR:", err) // don't fail — S3 already completed
	}

	return nil
}
