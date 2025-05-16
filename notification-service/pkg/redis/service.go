package redis

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClientInterface defines the methods of redis.Client used by RedisService
type RedisClientInterface interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	// Add other methods if used by RedisService in the future
}

type RedisService struct {
	Client RedisClientInterface
}

func NewRedisService(redisHost, redisPort, redisUsername, redisPassword string) (*RedisService, error) {

	client, err := RedisClient(redisHost, redisPort, redisUsername, redisPassword)

	if err != nil {
		return nil, err
	}
	log.Println("Initialized Redis repository.")
	return &RedisService{
		Client: client,
	}, nil
}

func (r *RedisService) Set(key string, data []byte, expiredTime time.Duration, ctx context.Context) error {
	err := r.Client.Set(ctx, key, data, expiredTime).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisService) Hset(key string, data string, expireAt time.Time, ctx context.Context) error {
	err := r.Client.Set(ctx, key, data, time.Until(expireAt)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisService) Get(key string, ctx context.Context) (string, error) {
	result, err := r.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", errors.New("key does not exist")
	} else if err != nil {
		return "", err
	}
	return result, nil
}

func (r *RedisService) Del(key string, ctx context.Context) error {
	_, err := r.Client.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}
