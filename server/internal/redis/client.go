package redis

import (
	"github.com/Shiwang0-0/HLS-Transcoder/server/internal/config"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	rDB *redis.Client
}

func NewRedisClient(config *config.AppConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: "",
	})
}
