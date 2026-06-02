package config

import "os"

const DefaultPartSize = 5 * 1024 * 1024

type AppConfig struct {
	BucketName string
	QueueURL   string
	RedisAddr  string
}

func LoadApp() *AppConfig {
	return &AppConfig{
		BucketName: os.Getenv("BUCKET_NAME"),
		QueueURL:   os.Getenv("QUEUE_URL"),
		RedisAddr:  os.Getenv("localhost:6379"),
	}
}
