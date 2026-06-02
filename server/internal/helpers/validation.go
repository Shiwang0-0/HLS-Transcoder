package helpers

import (
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/gofiber/fiber/v2"
)

func validationError(msg string) *fiber.Error {
	return &fiber.Error{
		Code:    fiber.StatusBadRequest,
		Message: msg,
	}
}

func ValidateData(data models.InitMultipartUploadRequest) *fiber.Error {
	allowedTypes := map[string]bool{
		"video/mp4": true,
	}

	if !allowedTypes[data.Type] {
		return validationError("Media type not allowed")
	}

	const maxSize = 500 * 1024 * 1024
	if data.Size > maxSize {
		return validationError("File size cannot exceed 500MB")
	}

	if data.Name == "" {
		return validationError("File name is required")
	}

	if data.Size <= 0 {
		return validationError("Invalid file size")
	}

	if data.Fingerprint == "" {
		return validationError("Fingerprint is required")
	}

	return nil
}
