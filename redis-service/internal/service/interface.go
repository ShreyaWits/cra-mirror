package service

type RedisServiceInterface interface {
	GetCache(key string) (string, error)
}
