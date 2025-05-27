package redis

import (
	"context"
	"time"
)

type RedisClient interface {
	Close() error
	SetCache(ctx context.Context, namespace, key, value string, ttl time.Duration) error
	GetCache(ctx context.Context, namespace, key string) (string, bool, error)
	InvalidateCache(ctx context.Context, namespace, key string) error
}