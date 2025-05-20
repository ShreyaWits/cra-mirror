package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClientInterface interface {
	Get(context.Context, string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(context.Context, ...string) *redis.IntCmd
}

type RedisInterface interface {
	GetCache(key string) (string, error)
	SetCache(key, value string, expiration *time.Duration) error
	InvalidateCache(key string) error
}
