package repository

import "time"

type RedisRepositoryInterface interface {
	GetCache(key string) (string, error)
	SetCache(key string, value string, ttl *time.Duration) error
	InvalidateCache(key string) error
}
