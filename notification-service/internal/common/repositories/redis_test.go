package repositories

import (
	"context"
	"encoding/json"
	"log"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"notification-service/internal/common/api/dtos"
	mock_repo "notification-service/internal/common/repositories/mock"
)

func TestNotificationRedis_GetNotificationConfig(t *testing.T) {
	// Setup mock Redis service
	redisService := &mock_repo.RedisService{}
	redisService.On("Get", mock.Anything, mock.Anything).Return(redis.NewStringResult(`[{"Service":"email","Primary":"sendgrid","Fallback":"smtp"}]`, nil), nil)
	// Create NotificationRedis instance
	notificationRedis := InitRedisRepo(redisService)

	// Test GetNotificationConfig
	configs, err := notificationRedis.GetNotificationConfig("notification-config")

	log.Println("configs", configs)
	log.Println("err", err)

	assert.NoError(t, err)

	// Unmarshal configs
	assert.NoError(t, err)

	// Verify configs
	assert.Equal(t, "email", configs[0].Service)
	assert.Equal(t, "sendgrid", configs[0].Primary)
	assert.Equal(t, "smtp", configs[0].Fallback)
}

func TestNotificationRedis_SetNotificationConfig(t *testing.T) {
	// Setup mock Redis service
	redisService := &mock_repo.RedisService{}

	configs := []dtos.ChannelConfig{
		{Service: "email", Primary: "sendgrid", Fallback: "smtp"},
	}
	expectedJSON, _ := json.Marshal(configs)

	redisService.On("Set", "notification-config", []byte(expectedJSON), mock.Anything, context.Background()).Return(nil)

	// Create NotificationRedis instance
	notificationRedis := InitRedisRepo(redisService)

	// Test SetNotificationConfig
	err := notificationRedis.SetNotificationConfig("notification-config", configs)

	assert.NoError(t, err)
	redisService.AssertExpectations(t)
}

func TestNotificationRedis_GetNotificationConfig_NilRedisService(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic due to nil Redis service, but none occurred")
		}
	}()

	// Create NotificationRedis instance with nil Redis service
	notificationRedis := InitRedisRepo(nil)

	// This should panic or error
	_, _ = notificationRedis.GetNotificationConfig("notification-config")
}

func TestNotificationRedis_SetNotificationConfig_NilRedisService(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic due to nil Redis service, but none occurred")
		}
	}()

	// Create NotificationRedis instance with nil Redis service
	notificationRedis := InitRedisRepo(nil)

	// Test SetNotificationConfig
	configs := []dtos.ChannelConfig{
		{Service: "email", Primary: "sendgrid", Fallback: "smtp"},
	}
	_ = notificationRedis.SetNotificationConfig("notification-config", configs)
}
