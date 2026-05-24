package config

import "os"

// this is application config, not the aws user config
type AppConfig struct {
	BucketName string
}

func Load() *AppConfig {
	return &AppConfig{
		BucketName: os.Getenv("BUCKET_NAME"),
	}
}
