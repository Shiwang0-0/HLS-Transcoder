package router

import (
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws/s3"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws/sqs"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/handlers"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/redis"
	"github.com/gofiber/fiber/v2"
)

func RouteSetup(app *fiber.App, s3Service *s3.Service, sqsService *sqs.Service, JobStore *redis.JobStore) {

	jobHandler := handlers.NewJobHandler(sqsService, JobStore)
	uploadHandler := handlers.NewUploadHandler(s3Service, jobHandler)

	api := app.Group("/api")
	api.Post("/presigned-url", uploadHandler.GeneratePresignedURL)
	api.Post("/notify-upload", uploadHandler.NotifyUpload)
	api.Get("/job/:jobid", jobHandler.GetJob)
}
