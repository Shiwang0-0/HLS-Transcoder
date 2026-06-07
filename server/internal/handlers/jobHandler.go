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
	response["job"] = job // already public

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
		response["msg"] = "Error creating job"
	}

	if err := h.JobService.UploadToSQS(c.Context(), job); err != nil {
		// even if failed, return the job the retry go routine will eventually push to SQS
		response["msg"] = "Video uploaded and queued for processing."
		fmt.Printf("Changing Job: %s status to failed_to_queue", job.JobID)

		response["job"] = job.ToPublic()
		return c.Status(202).JSON(response)
	}

	response["job"] = job.ToPublic()

	return c.Status(200).JSON(response)
}

func (h *JobHandler) GetAllJobs(c *fiber.Ctx) error {
	response := fiber.Map{
		"msg": "All jobs fetched",
	}
	jobs, err := h.JobService.GetAllJobs(c.Context())
	if err != nil {
		response["msg"] = "Error fetching jobs"
		return c.Status(500).JSON(response)
	}
	response["jobs"] = jobs // already public
	return c.Status(200).JSON(response)
}
