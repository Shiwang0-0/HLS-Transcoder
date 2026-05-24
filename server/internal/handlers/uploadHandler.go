package handlers

import (
	"fmt"

	mys3 "github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws/s3"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/gofiber/fiber/v2"
)

type UploadHandler struct {
	S3Service  *mys3.Service
	JobHandler *JobHandler // abstracting the sqs handler
}

func NewUploadHandler(s3Service *mys3.Service, jobHandler *JobHandler) *UploadHandler {
	return &UploadHandler{
		S3Service:  s3Service,
		JobHandler: jobHandler,
	}
}

func (h *UploadHandler) GeneratePresignedURL(c *fiber.Ctx) error {

	context := fiber.Map{
		"msg": "PresignedURL generated",
	}

	var metadata models.VideoMetadata

	if err := c.BodyParser(&metadata); err != nil {
		context["msg"] = "Error parsing request body"
		return c.Status(400).JSON(context)
	}

	validationErr := validateMetadata(metadata)
	if validationErr != nil {
		context["msg"] = validationErr.Message
		return c.Status(400).JSON(context)
	}

	url, err := h.S3Service.GeneratePresignedURL(c.Context(), metadata)
	if err != nil {
		context["msg"] = "Error generating Presigned url"
		return c.Status(400).JSON(context)
	}
	fmt.Println("Presigned URL: ", url)
	context["url"] = url
	context["key"] = metadata.Name

	return c.Status(200).JSON(context)
}
func validateMetadata(metadata models.VideoMetadata) *fiber.Error {
	allowedTypes := map[string]bool{
		"video/mp4": true,
	}
	if !allowedTypes[metadata.Type] {
		return &fiber.Error{Message: "Media type not allowed"}
	}

	// validate file size (500MB limit)
	const maxSize = 500 * 1024 * 1024
	if metadata.Size > maxSize {
		return &fiber.Error{Message: "File size cannot exceed 500MB"}
	}

	return nil
}

func (h *UploadHandler) NotifyUpload(c *fiber.Ctx) error {
	context := fiber.Map{
		"msg": "Service Notified",
	}

	var data models.NotifyData

	if err := c.BodyParser(&data); err != nil {
		context["msg"] = "Error parsing request body"
		return c.Status(400).JSON(context)
	}

	// push an entry into sqs regarding the s3 upload
	err := h.JobHandler.PutInQueue(c.Context(), data)
	if err != nil {
		context["msg"] = "Error adding to sqs"
		return c.Status(400).JSON(context)
	}

	// fmt.Println("notify data: ", data.Key)

	return c.Status(200).JSON(context)
}
