package router

import (
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws/s3"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/handlers"
	"github.com/gofiber/fiber/v2"
)

func RouteSetup(app *fiber.App, s3Service *s3.Service) {

	uploadHandler := handlers.NewUploadHandler(s3Service)

	api := app.Group("/api")
	api.Post("/presigned-url", uploadHandler.GeneratePresignedURL)
}
