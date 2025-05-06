package database

import (
	"context"
	"fmt"
	"log"
	configEnv "protected_link/internal/configs"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Client *redis.Client
	Ctx    context.Context
}

// ConnectRedis initializes and returns a Redis client using app config.
func ConnectRedis(cfg *configEnv.Config) (*RedisConfig, error) {
	ctx := context.Background()
	var client *redis.Client
	var err error

	options := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.DBHost, cfg.DBPort),
		Password: cfg.DBPassword,
		DB:       0,
	}

	maxRetries := 3
	retryDelay := 2 * time.Second

	for i := 1; i <= maxRetries; i++ {
		log.Printf("🔄 Attempting to connect to Redis (Attempt %d/%d)...", i, maxRetries)
		client = redis.NewClient(options)

		_, err = client.Ping(ctx).Result()
		if err == nil {
			log.Println("✅ Redis connection established successfully")
			return &RedisConfig{
				Client: client,
				Ctx:    ctx,
			}, nil
		}

		log.Printf("❌ Redis connection failed (Attempt %d/%d): %v", i, maxRetries, err)
		if i < maxRetries {
			log.Printf("⏳ Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	// After max retries, return an error
	log.Printf("🔻 Unable to connect to Redis after %d attempts.", maxRetries)
	return nil, fmt.Errorf("failed to connect to Redis after %d retries: %w", maxRetries, err)
}

// Close closes the Redis connection
func (r *RedisConfig) Close() {
	if r.Client != nil {
		if err := r.Client.Close(); err != nil {
			log.Printf("❌ Error closing Redis connection: %v", err)
		} else {
			log.Println("✅ Redis connection closed")
		}
	}
}
