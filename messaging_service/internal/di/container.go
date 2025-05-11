package di

import (
	"messaging_service/internal/config"
	"messaging_service/internal/messaging_service/handler"
	"messaging_service/pkg/kafka"
	"messaging_service/pkg/logger"
)

// Container holds all the dependencies for the application
type Container struct {
	// Config
	Config *config.Config

	// Key Manager
	MessagingHandler *handler.MessagingHandler

	Kafka *kafka.KafkaPkg
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
	msgHandler := handler.NewMessagingHandler(cfg)
	kafka := &kafka.KafkaPkg{}

	// Build container
	container := &Container{
		Config:           cfg,
		MessagingHandler: msgHandler,
		Kafka:            kafka,
	}

	return container, nil
}
