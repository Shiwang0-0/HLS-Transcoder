package handlers

import (
	"log"
	"strconv"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/helpers"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/service"
	"github.com/gofiber/fiber/v2"
)

type UploadHandler struct {
	UploadService *service.UploadService
	JobService    *service.JobService
}

func NewUploadHandler(uploadService *service.UploadService, jobService *service.JobService) *UploadHandler {
	return &UploadHandler{
		UploadService: uploadService,
		JobService:    jobService,
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

	if session.IsNewSession {
		jobID, jobErr := h.JobService.CreateJobShell(c.Context(), session.VideoID, session.Key)
		if jobErr != nil {
			log.Printf("failed to pre-create job shell for videoID %s: %v", session.VideoID, jobErr)
		}
		session.JobID = jobID
	} else if existing, jobErr := h.JobService.GetJobByVideoID(c.Context(), session.VideoID); jobErr == nil {
		session.JobID = existing.JobID
	}

	response["session"] = session
	return c.Status(200).JSON(response)
}

func (h *UploadHandler) VerifyAndPersistParts(c *fiber.Ctx) error {
	response := fiber.Map{
		"msg": "Parts verfied and persisted",
	}
	sessionIDStr := c.Params("sessionId")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid session id"})
	}

	// fmt.Println("sessionID: ", sessionID)

	var data []models.Part
	if err := c.BodyParser(&data); err != nil {
		response["msg"] = "Error parsing request body"
		return c.Status(400).JSON(response)
	}

	verifiedParts, missingParts, err := h.UploadService.VerifyAndPersistParts(c.Context(), sessionID, data)
	if err != nil {
		response["msg"] = "Error in parts verification and persistence"
		return c.Status(500).JSON(response)
	}

	response["verifiedParts"] = verifiedParts
	response["missingParts"] = missingParts

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

	session, err := h.UploadService.GetSessionByID(c.Context(), data.SessionID)
	if err != nil || session == nil {
		response["msg"] = "Session not found"
		return c.Status(400).JSON(response)
	}

	if err := h.UploadService.CompleteMultipartUpload(c.Context(), data); err != nil {
		response["msg"] = "Error completing multipart upload"
		return c.Status(500).JSON(response)
	}

	if err := h.JobService.MarkUploaded(c.Context(), session.VideoID); err != nil {
		log.Printf("failed to mark job uploaded for videoID %s: %v", session.VideoID, err)
	}

	return c.Status(200).JSON(response)
}
