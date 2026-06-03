package handlers

import (
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/helpers"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/service"
	"github.com/gofiber/fiber/v2"
)

type UploadHandler struct {
	UploadService *service.UploadService
}

func NewUploadHandler(uploadService *service.UploadService) *UploadHandler {
	return &UploadHandler{
		UploadService: uploadService,
	}
}

func (h *UploadHandler) GeneratePresignedPartURL(c *fiber.Ctx) error {
	response := fiber.Map{
		"msg": "Presigned URL generated",
	}

	var data models.PresignedPartURLRequest
	if err := c.BodyParser(&data); err != nil {
		response["msg"] = "Error parsing request body"
		return c.Status(400).JSON(response)
	}

	urlResponse, err := h.UploadService.GeneratePresignedPartURL(c.Context(), data)
	if err != nil {
		response["msg"] = "Error generating presigned URL"
		return c.Status(500).JSON(response)
	}
	response["url"] = urlResponse.URL
	return c.Status(200).JSON(response)
}

func (h *UploadHandler) InitMultipartUpload(c *fiber.Ctx) error {
	response := fiber.Map{
		"msg": "Multipart upload initialized",
	}

	var data models.InitMultipartUploadRequest

	if err := c.BodyParser(&data); err != nil {
		response["msg"] = "Error parsing request body"
		return c.Status(400).JSON(response)
	}

	if err := helpers.ValidateData(data); err != nil {
		response["msg"] = err.Error()
		return c.Status(400).JSON(response)
	}
	session, err := h.UploadService.InitMultipartUpload(c.Context(), data)
	if err != nil {
		response["msg"] = "Error initializing multipart upload"
		return c.Status(500).JSON(response)
	}
	response["session"] = session
	return c.Status(200).JSON(response)
}

func (h *UploadHandler) CompleteMultipartUpload(c *fiber.Ctx) error {
	response := fiber.Map{
		"msg": "Multipart upload completed",
	}
	var data models.CompleteMultipartUploadRequest
	if err := c.BodyParser(&data); err != nil {
		response["msg"] = "Error parsing request body"
		return c.Status(400).JSON(response)
	}

	if err := h.UploadService.CompleteMultipartUpload(c.Context(), data); err != nil {
		response["msg"] = "Failed to complete multipart upload"
		return c.Status(500).JSON(response)
	}

	response["videoID"] = data.VideoID
	response["uploadID"] = data.UploadID
	return c.Status(200).JSON(response)
}
