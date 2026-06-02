package main

import (
	"context"
	"log"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/config"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/repository"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/service"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/worker"
	"github.com/joho/godotenv"
)

func main() {

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
	s3Client := aws.NewS3Client(awsConfig)
	sqsClient := aws.NewSQSClient(awsConfig)

	db, err := config.LoadMySQL()
	if err != nil {
		log.Fatal("MySQL config failed", err)
	}

	// initialize services
	s3Service := service.NewS3Service(s3Client, appConfig.BucketName)
	sqsService := service.NewSQSService(sqsClient, appConfig.QueueURL)
	jobRepository := repository.NewJobRepository(db)

	w := worker.NewWorker(s3Service, sqsService, jobRepository)

	w.Start(ctx) // start polling
}
