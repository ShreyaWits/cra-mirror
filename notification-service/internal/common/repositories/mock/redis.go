// notification-service/internal/common/repositories/mock/redis.go

package mock

import (
	"context"
	redis_service "notification-service/pkg/redis"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
)

type RedisService struct {
	mock.Mock
	redis_service.RedisServiceInterface
}

func (m *RedisService) Get(key string, ctx context.Context) (string, error) {
	args := m.Called(key, ctx)
	cmd := args.Get(0).(*redis.StringCmd)
	return cmd.Result()
}

func (m *RedisService) Set(key string, value []byte, ttl time.Duration, ctx context.Context) error {
	args := m.Called(key, value, ttl, ctx)
	return args.Error(0)
}
