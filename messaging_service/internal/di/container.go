package di

import (
	"messaging_service/internal/config"
	"messaging_service/internal/messaging_service/handler"
	"messaging_service/pkg/logger"
)

// Container holds all the dependencies for the application
type Container struct {
	// Config
	Config *config.Config

	// Key Manager
	MessagingHandler *handler.MessagingHandler
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

	// Build container
	container := &Container{
		Config:           cfg,
		MessagingHandler: msgHandler,
	}

	return container, nil
}
