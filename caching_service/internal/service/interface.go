package service

import "redis-service/internal/dto"

type RedisServiceInterface interface {
	GetCache(payload *dto.GetCacheRequest) (string, error)
	SetCache(payload *dto.SetCacheRequest) (bool, error)
	InvalidateCache(payload *dto.DeleteCacheRequest) (bool, error)
}
