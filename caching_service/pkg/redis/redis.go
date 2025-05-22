package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client RedisClientInterface
}

func NewRedisService(URL, USERNAME, PASSWORD string, DB int) *RedisClient {
	// Initialize Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     URL,
		Password: PASSWORD,
		DB:       DB,
		Username: USERNAME,
	})

	// wrap the client in a struct
	return &RedisClient{
		client: client,
	}
}

// RedisServiceInterface defines the methods for interacting with Redis
func (c *RedisClient) GetCache(key string) (string, error) {
	val, err := c.client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return val, nil
}

func (c *RedisClient) SetCache(key, value string, expiration *time.Duration) error {
	if expiration == nil {
		exp := 0 * time.Second
		expiration = &exp
	}
	err := c.client.Set(context.Background(), key, value, *expiration).Err()
	return err
}

func (c *RedisClient) InvalidateCache(key string) error {
	err := c.client.Del(context.Background(), key).Err()
	return err
}
