package service

import (
	"context"
	"notification-service/internal/common/api/dtos"
	"notification-service/internal/common/models"
)

type NotificationRedisInterface interface {
	GetNotificationConfig(key string) ([]dtos.ChannelConfig, error)
	SetNotificationConfig(key string, config []dtos.ChannelConfig) error
}
type ConfigRepositoryInterface interface {
	GetConfig() ([]dtos.ChannelConfig, error)
}
type NotificationRepositoryInterface interface {
	SaveNotification(n *models.Notification) (*string, error)
	UpdateNotificationByID(id string, n *models.Notification) error
}
type TemporalClientInterface interface {
	ExecuteEmailWorkflow(data map[string]interface{}) map[string]interface{}
	ExecuteSmsWorkflow(data map[string]interface{}) map[string]interface{}
	ExecutePushNotificationWorkflow(data map[string]interface{}) map[string]interface{}
	ExecuteWhatsappWorkflow(data map[string]interface{}) map[string]interface{}
}
type KafkaPublisherInterface interface {
	Publish(ctx context.Context, topic string, key, value []byte) error
}
