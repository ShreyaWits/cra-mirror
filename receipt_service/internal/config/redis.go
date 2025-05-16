package config

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	ctx         = context.Background()
)

func InitRedis(addr string) {
	redisClient = redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

func GetRedisClient() *redis.Client {
	return redisClient
}
