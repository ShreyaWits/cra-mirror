package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"thirdparty_service/internal/models"
	"time"

	"github.com/redis/go-redis/v9"
)

type WebhookRepository interface {
	SaveWebhook(key string, data *models.Webhook) error
}

type RedisRepo struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisRepo(addr string) *RedisRepo {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: addr, // e.g., "localhost:6379"
		DB:   0,
	})

	// Optional: check connection
	if err := client.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}

	return &RedisRepo{
		client: client,
		ctx:    ctx,
	}
}

func (r *RedisRepo) SaveWebhook(key string, data *models.Webhook) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook: %w", err)
	}

	redisKey := fmt.Sprintf("webhook:%s", key)
	err = r.client.Set(r.ctx, redisKey, jsonData, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to save to Redis: %w", err)
	}

	return nil
}
