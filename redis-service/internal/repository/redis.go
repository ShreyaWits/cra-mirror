package repository

import (
	"redis-service/pkg/redis"
	"time"
)

type RedisRepo struct {
	db redis.RedisInterface
}

func NewRedisRepository() *RedisRepo {
	// Initialize Redis client
	client := redis.NewRedisService("localhost:6379", "password", 0)
	return &RedisRepo{
		db: client,
	}
}

func (r *RedisRepo) GetCache(key string) (string, error) {
	return r.db.GetCache(key)
}
func (r *RedisRepo) SetCache(key string, value string, ttl *time.Duration) error {
	return r.db.SetCache(key, value, ttl)
}
func (r *RedisRepo) InvalidateCache(key string) error {
	return r.db.InvalidateCache(key)
}
