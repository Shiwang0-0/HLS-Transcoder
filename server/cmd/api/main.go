package main

import (
	"context"
	"log"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws/s3"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/config"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/router"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {

	// creating an http server
	app := fiber.New()
	app.Use(logger.New())

	app.Use(config.CorsConfig)

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env")
	}

	// loading app config
	appConfig := config.Load()

	ctx := context.Background()

	// initialize s3 client
	s3Client, err := s3.NewS3Client(ctx)

	if err != nil {
		log.Fatal(err)
	}

	// initialize s3 service
	s3Service := s3.NewService(s3Client, appConfig.BucketName)

	router.RouteSetup(app, s3Service)

	app.Listen(":8000")
}
