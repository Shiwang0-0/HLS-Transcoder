package main

import (
	"context"
	"log"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws/s3"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws/sqs"
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

	ctx := context.Background()

	// app config
	appConfig := config.LoadApp()

	// aws sdk config
	awsConfig, err := config.LoadAWS(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// initialize clients
	s3Client := s3.NewS3Client(awsConfig)
	sqsClient := sqs.NewSQSClient(awsConfig)

	if err != nil {
		log.Fatal(err)
	}

	// initialize services
	s3Service := s3.NewService(s3Client, appConfig.BucketName)
	sqsService := sqs.NewService(sqsClient, appConfig.QueueURL)

	// initlaize job handler

	router.RouteSetup(app, s3Service, sqsService)

	app.Listen(":8000")
}
