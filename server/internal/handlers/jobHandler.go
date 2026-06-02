package handlers

import (
	"fmt"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/service"
	"github.com/gofiber/fiber/v2"
)

type JobHandler struct {
	JobService *service.JobService
}

func NewJobHandler(jobService *service.JobService) *JobHandler {
	return &JobHandler{
		JobService: jobService,
	}
}

func (h *JobHandler) GetJob(c *fiber.Ctx) error {
	response := fiber.Map{
		"msg": "Job found",
	}
	jobID := c.Params("jobid")

	job, err := h.JobService.GetJob(c.Context(), jobID)
	if err != nil {
		response["msg"] = "Error searching for Job"
		return c.Status(500).JSON(response)
	}
	if job == nil {
		response["msg"] = "Job not found"
		return c.Status(404).JSON(response)
	}
	response["job"] = job

	return c.Status(200).JSON(response)
}

func (h *JobHandler) CreateTranscodingJob(c *fiber.Ctx) error {
	response := fiber.Map{
		"msg": "Service notified",
	}

	var data models.JobCreationRequest
	if err := c.BodyParser(&data); err != nil {
		response["msg"] = "Error parsing request body"
		return c.Status(400).JSON(response)
	}

	job, err := h.JobService.CreateTranscodingJob(c.Context(), data)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	if err := h.JobService.UploadToSQS(c.Context(), job); err != nil {
		response["msg"] = "Error notifying upload"
		return c.Status(500).JSON(response)
	}

	response["job"] = job

	return c.Status(200).JSON(response)
}
