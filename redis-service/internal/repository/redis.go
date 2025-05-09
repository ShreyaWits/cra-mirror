package repository

import (
	"redis-service/pkg/redis"
	"time"
)

type RedisRepo struct {
	db redis.RedisInterface
}

func NewRedisRepository(client redis.RedisInterface) *RedisRepo {
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
