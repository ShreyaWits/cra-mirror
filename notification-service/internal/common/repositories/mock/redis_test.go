// notification-service/internal/common/repositories/mock/redis_test.go

package mock

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisService_Get(t *testing.T) {
	redisService := &RedisService{}
	redisService.On("Get", "key", context.Background()).Return(redis.NewStringResult("value", nil))

	result, err := redisService.Get("key", context.Background())
	assert.Equal(t, "value", result)
	assert.Equal(t, nil, err)
}

func TestRedisService_Set(t *testing.T) {
	redisService := &RedisService{}
	redisService.On("Set", "key", []byte("value"), time.Hour, context.Background()).Return(nil)

	err := redisService.Set("key", []byte("value"), time.Hour, context.Background())
	assert.Equal(t, nil, err)
}
