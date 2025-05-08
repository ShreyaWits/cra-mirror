package service

import "redis-service/internal/repository"

type RedisService struct {
	repo repository.RedisRepositoryInterface
}

func NewRedisService(redisRepo repository.RedisRepositoryInterface) *RedisService {
	return &RedisService{
		repo: redisRepo,
	}
}
func (s *RedisService) GetCache(key string) (string, error) {
	return s.repo.GetCache(key)
}
