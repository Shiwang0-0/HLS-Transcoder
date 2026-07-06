package main

import (
	"context"
	"log"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/config"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/repository"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/router"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {

	// creating an http server
	app := fiber.New()
	app.Use(logger.New())

	app.Use(config.CorsConfig)

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	ctx := context.Background()

	// app config
	appConfig := config.LoadApp()

	// aws sdk config
	awsConfig, err := config.LoadAWS(ctx)
	if err != nil {
		log.Fatal("AWS config failed", err)
	}

	// db initalize
	db, err := config.LoadMySQL()
	if err != nil {
		log.Fatal("MySQL config failed", err)
	}

	// initialize clients
	s3Client := aws.NewS3Client(awsConfig)
	sqsClient := aws.NewSQSClient(awsConfig)

	// initialize repositories
	jobRepository := repository.NewJobRepository(db)
	uploadRepository := repository.NewUploadRepository(db)

	// initialize services
	s3Service := service.NewS3Service(s3Client, appConfig.BucketName)
	sqsService := service.NewSQSService(sqsClient, appConfig.QueueURL)
	jobService := service.NewJobService(jobRepository, sqsService)
	uploadService := service.NewUploadService(uploadRepository, s3Service)

	router.RouteSetup(app, s3Service, sqsService, jobService, uploadService, db)

	// searched for jobs that were failed to be pushed to queue
	go jobService.StartRetrySweeper(context.Background())

	app.Listen(":8000")
}
