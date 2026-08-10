package router

import (
	"database/sql"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/handlers"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/service"
	"github.com/gofiber/fiber/v2"
)

func RouteSetup(app *fiber.App, s3Service *service.S3Service, sqsService *service.SQSService, jobService *service.JobService, uploadService *service.UploadService, db *sql.DB) {

	jobHandler := handlers.NewJobHandler(jobService)
	uploadHandler := handlers.NewUploadHandler(uploadService, jobService)

	api := app.Group("/api")

	api.Post("/presigned-part-url", uploadHandler.GeneratePresignedPartURL)
	api.Post("/init-multipart-upload", uploadHandler.InitMultipartUpload)
	api.Post("/complete-multipart-upload", uploadHandler.CompleteMultipartUpload)
	api.Patch("/uploads/:sessionId/parts", uploadHandler.VerifyAndPersistParts)

	api.Get("/jobs", jobHandler.GetAllJobs)
	api.Get("/job/:jobid", jobHandler.GetJob)
	app.Get("/api/job/:jobid/stream", jobHandler.StreamJobStatus)
	api.Post("/job", jobHandler.QueueTranscodingJob)
}
