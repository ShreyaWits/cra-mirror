package repository

import (
	"log/slog"
	"redis-service/pkg/redis"
	"time"
)

type RedisRepo struct {
	db     redis.RedisInterface
	logger *slog.Logger
}

func NewRedisRepository(client redis.RedisInterface, logger *slog.Logger) *RedisRepo {
	return &RedisRepo{
		db:     client,
		logger: logger,
	}
}

func (r *RedisRepo) GetCache(key string) (string, error) {
	r.logger.Debug("getting cache from redis", "key", key)
	return r.db.GetCache(key)
}
func (r *RedisRepo) SetCache(key string, value string, ttl *time.Duration) error {
	r.logger.Debug("setting cache in redis", "key", key, "ttl", ttl)
	return r.db.SetCache(key, value, ttl)
}
func (r *RedisRepo) InvalidateCache(key string) error {
	r.logger.Debug("invalidating cache in redis", "key", key)
	return r.db.InvalidateCache(key)
}

