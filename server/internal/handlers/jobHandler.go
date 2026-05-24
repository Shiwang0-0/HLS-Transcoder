package handlers

import (
	"context"

	mysqs "github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws/sqs"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/models"
)

type JobHandler struct {
	SQSService *mysqs.Service
}

func NewJobHandler(sqsService *mysqs.Service) *JobHandler {
	return &JobHandler{SQSService: sqsService}
}

func (h *JobHandler) PutInQueue(ctx context.Context, data models.NotifyData) error {
	return h.SQSService.PutInQueue(ctx, data)
}
