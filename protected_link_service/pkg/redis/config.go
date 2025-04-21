package database

import (
	"context"
	"fmt"
	"log"
	configEnv "protected_link/internal/configs"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Client *redis.Client
	Ctx    context.Context
}

// ConnectRedis initializes and returns a Redis client using app config.
func ConnectRedis(cfg *configEnv.Config) (*RedisConfig, error) {
	ctx := context.Background()

	options := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.DBHost, cfg.DBPort),
		Password: cfg.DBPassword,
		DB:       0,
	}

	client := redis.NewClient(options)

	// Ping with context to verify connection
	if _, err := client.Ping(ctx).Result(); err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
		return nil, err
	}

	log.Println("✅ Redis connection established successfully")

	return &RedisConfig{
		Client: client,
		Ctx:    ctx,
	}, nil
}
