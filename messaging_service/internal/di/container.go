package di

import (
	"messaging_service/internal/config"
	"messaging_service/internal/messaging_service/handler"
	"messaging_service/internal/messaging_service/service"
	"messaging_service/pkg/kafka"
	"messaging_service/pkg/logger"
)

// Container holds all the dependencies for the application
type Container struct {
	// Config
	Config *config.Config

	// Key Manager
	MessagingHandler *handler.MessagingHandler

	Producer *kafka.Producer
}

// NewContainer creates a new dependency injection container
func NewContainer() (*Container, error) {
	// Initialize logger
	logger.InitLogger()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Initialize dependencies
	producer := kafka.NewProducer()
	admin := kafka.NewAdmin(kafka.KafkaConfig{
		Brokers: cfg.KafkaBrokers,
	})

	messagingService := service.NewMessagingService(producer, admin)
	msgHandler := handler.NewMessagingHandler(cfg, messagingService)

	// Build container
	container := &Container{
		Config:           cfg,
		MessagingHandler: msgHandler,
		Producer:         producer,
	}

	return container, nil
}
