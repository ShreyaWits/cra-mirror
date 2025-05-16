package repositories

import (
	"context"
	"encoding/json"
	"log"
	"notification-service/internal/common/api/dtos"

	redis_service "notification-service/pkg/redis"
	"time"
)

type RedisRepositoryInterface interface {
	GetNotificationConfig(key string) ([]dtos.ChannelConfig, error)
	SetNotificationConfig(key string, config []dtos.ChannelConfig) error
}

type NotificationRedis struct {
	redis redis_service.RedisServiceInterface
}

func InitRedisRepo(redisService redis_service.RedisServiceInterface) *NotificationRedis {
	return &NotificationRedis{
		redis: redisService,
	}
}

func (n *NotificationRedis) GetNotificationConfig(key string) ([]dtos.ChannelConfig, error) {
	var configs []dtos.ChannelConfig
	configuration, err := n.redis.Get(key, context.Background())
	if err != nil {
		return nil, err
	}
	if configuration == "" {
		return nil, nil
	}

	// un marshall
	configErr := json.Unmarshal([]byte(configuration), &configs)
	if configErr != nil {
		log.Println("Error unmarshaling configuration:", configErr.Error())
		return nil, configErr
	}

	return configs, nil
}

func (n *NotificationRedis) SetNotificationConfig(key string, configs []dtos.ChannelConfig) error {
	config, err := json.Marshal(configs)
	if err != nil {
		log.Println("Error marshaling configuration:", err.Error())
		return err
	}
	err = n.redis.Set(key, []byte(config), time.Duration(24*time.Hour), context.Background())
	if err != nil {
		log.Println("Error setting configuration in Redis:", err.Error())
		return err
	}
	return nil
}
