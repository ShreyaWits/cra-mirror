package service

import (
	"redis-service/internal/dto"
	"redis-service/internal/repository"
	"redis-service/internal/utils"
	"time"
)

type RedisService struct {
	repo repository.RedisRepositoryInterface
}

func NewRedisService(redisRepo repository.RedisRepositoryInterface) *RedisService {
	return &RedisService{
		repo: redisRepo,
	}
}
func (s *RedisService) GetCache(payload *dto.GetCacheRequest) (string, error) {
	key := utils.GenerateRedisCacheKey(payload.Namespace, payload.Key)
	return s.repo.GetCache(key)
}
func (s *RedisService) SetCache(payload *dto.SetCacheRequest) (bool, error) {
	key := utils.GenerateRedisCacheKey(payload.Namespace, payload.Key)

	var ttl *time.Duration
	if payload.TTL > 0 {
		ttl = new(time.Duration)
		*ttl = time.Duration(payload.TTL) * time.Second
	}

	err := s.repo.SetCache(key, payload.Value, ttl)
	if err != nil {
		return false, err
	}
	// If the cache is set successfully, return true
	return true, nil
}

func (s *RedisService) InvalidateCache(payload *dto.DeleteCacheRequest) (bool, error) {
	key := utils.GenerateRedisCacheKey(payload.Namespace, payload.Key)
	// Invalidate the cache using the repository
	err := s.repo.InvalidateCache(key)
	if err != nil {
		return false, err
	}
	return true, nil
}
