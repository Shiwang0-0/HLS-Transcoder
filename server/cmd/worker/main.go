package main

import (
	"context"
	"log"

	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws/s3"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/aws/sqs"
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/config"
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
	s3Client := s3.NewS3Client(awsConfig)
	sqsClient := sqs.NewSQSClient(awsConfig)

	if err != nil {
		log.Fatal(err)
	}

	// initialize services
	s3Service := s3.NewService(s3Client, appConfig.BucketName)
	sqsService := sqs.NewService(sqsClient, appConfig.QueueURL)

	w := worker.NewWorker(s3Service, sqsService)

	w.Start(ctx) // start polling
}
