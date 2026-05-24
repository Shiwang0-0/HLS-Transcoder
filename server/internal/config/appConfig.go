package config

import "os"

type AppConfig struct {
	BucketName string
	QueueURL   string
}

func LoadApp() *AppConfig {
	return &AppConfig{
		BucketName: os.Getenv("BUCKET_NAME"),
		QueueURL:   os.Getenv("QUEUE_URL"),
	}
}
