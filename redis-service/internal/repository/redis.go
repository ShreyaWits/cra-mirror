package repository

import "redis-service/pkg/redis"

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
