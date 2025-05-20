// redis_interface.go
package database

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// IRedisClient abstracts Redis operations.
type IRedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Ping(ctx context.Context) *redis.StatusCmd
	Close() error
}
