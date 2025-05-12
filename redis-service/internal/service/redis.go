package service

import (
	"log/slog"
	"redis-service/internal/dto"
	"redis-service/internal/repository"
	"redis-service/internal/utils"
	"time"
)

type RedisService struct {
	repo   repository.RedisRepositoryInterface
	logger *slog.Logger
}

func NewRedisService(redisRepo repository.RedisRepositoryInterface, logger *slog.Logger) *RedisService {
	return &RedisService{
		repo:   redisRepo,
		logger: logger,
	}
}
func (s *RedisService) GetCache(payload *dto.GetCacheRequest) (string, error) {
	s.logger.Info("getting cache", "namespace", payload.Namespace, "key", payload.Key)
	key := utils.GenerateRedisCacheKey(payload.Namespace, payload.Key)
	return s.repo.GetCache(key)
}
func (s *RedisService) SetCache(payload *dto.SetCacheRequest) (bool, error) {
	s.logger.Info("setting cache", "namespace", payload.Namespace, "key", payload.Key, "ttl", payload.TTL)
	key := utils.GenerateRedisCacheKey(payload.Namespace, payload.Key)

	var ttl *time.Duration
	if payload.TTL > 0 {
		ttl = new(time.Duration)
		*ttl = time.Duration(payload.TTL) * time.Second
	}

	err := s.repo.SetCache(key, payload.Value, ttl)
	if err != nil {
		s.logger.Error("failed to set cache", "error", err, "key", key)
		return false, err
	}
	// If the cache is set successfully, return true
	s.logger.Info("cache set successfully", "key", key)
	return true, nil
}

func (s *RedisService) InvalidateCache(payload *dto.DeleteCacheRequest) (bool, error) {
	s.logger.Info("invalidating cache", "namespace", payload.Namespace, "key", payload.Key)
	key := utils.GenerateRedisCacheKey(payload.Namespace, payload.Key)
	// Invalidate the cache using the repository
	err := s.repo.InvalidateCache(key)
	if err != nil {
		s.logger.Error("failed to invalidate cache", "error", err, "key", key)
		return false, err
	}
	s.logger.Info("cache invalidated successfully", "key", key)
	return true, nil
}
