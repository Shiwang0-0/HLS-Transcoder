package handlers

import (
	"context"

	mysqs "github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws/sqs"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/redis"
	"github.com/gofiber/fiber/v2"
)

type JobHandler struct {
	SQSService *mysqs.Service
	JobStore   *redis.JobStore
}

func NewJobHandler(sqsService *mysqs.Service, jobStore *redis.JobStore) *JobHandler {
	return &JobHandler{SQSService: sqsService, JobStore: jobStore}
}

func (h *JobHandler) PutInQueue(ctx context.Context, data models.NotifyData) error {

	job := models.JobStatus{
		JobID:    data.JobID,
		Status:   "queued",
		Stage:    "sqs",
		Progress: 0,
		Key:      data.Key,
	}

	err := h.JobStore.SetJob(ctx, job)
	if err != nil {
		return err
	}

	return h.SQSService.PutInQueue(ctx, data)
}

func (h *JobHandler) GetJob(c *fiber.Ctx) error {
	jobID := c.Params("jobid")

	job, err := h.JobStore.GetJob(c.Context(), jobID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "job not found",
		})
	}

	return c.JSON(job)
}
