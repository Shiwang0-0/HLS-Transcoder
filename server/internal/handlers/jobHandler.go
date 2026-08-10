package handlers

import (
	"bufio"
	"encoding/json"
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

func (h *JobHandler) StreamJobStatus(c *fiber.Ctx) error {
	jobID := c.Params("jobid")
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	ctx := c.Context()
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		updates := h.JobService.WatchJob(ctx, jobID)

		for update := range updates {
			if update.Err != nil {
				fmt.Fprintf(w, "event: error\ndata: %s\n\n", update.Err.Error())
				w.Flush()
				return
			}

			payload, _ := json.Marshal(fiber.Map{
				"status": update.Status,
				"stage":  update.Stage,
			})
			fmt.Fprintf(w, "data: %s\n\n", payload)

			if err := w.Flush(); err != nil {
				// client disconnected
				// not context cancellation, so this is the actual disconnect signal
				return
			}
		}
	})

	return nil

}

func (h *JobHandler) QueueTranscodingJob(c *fiber.Ctx) error {
	response := fiber.Map{
		"msg": "Service notified",
	}

	var data models.JobCreationRequest
	if err := c.BodyParser(&data); err != nil {
		response["msg"] = "Error parsing request body"
		return c.Status(400).JSON(response)
	}

	job, err := h.JobService.QueueTranscodingJob(c.Context(), data)
	if err != nil {
		response["msg"] = err.Error()
		return c.Status(409).JSON(response)
	}

	if err := h.JobService.UploadToSQS(c.Context(), job); err != nil {
		// even if failed, the retry goroutine will eventually push to SQS
		fmt.Printf("Changing Job: %s status to failed_to_queue\n", job.JobID)
		response["msg"] = "Video uploaded and queued for processing."
		response["job"] = job
		return c.Status(202).JSON(response)
	}

	response["msg"] = "Video queued for processing."
	response["job"] = job
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
