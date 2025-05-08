package repository

type RedisRepositoryInterface interface {
	GetCache(key string) (string, error)
}
