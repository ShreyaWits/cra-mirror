package di

import (
	"messaging_service/internal/config"
	"messaging_service/internal/messaging_service/handler"
	"messaging_service/internal/messaging_service/service"
	"messaging_service/pkg/logger"
)

// Container holds all the dependencies for the application
type Container struct {
	// Config
	Config *config.Config

	// Service handlers
	MessagingHandler *handler.MessagingHandler

	// Services
	MessagingService service.MessagingService
}

// NewContainer creates a new dependency injection container
//
// This initializes the application dependencies:
//   - Creates a logger
//   - Loads configuration
//   - Sets up the Confluent Kafka-based messaging service for reliable message processing
//     with exactly-once delivery guarantees, transaction support, and automatic retries
//   - Creates the messaging service handler
func NewContainer() (*Container, error) {
	// Initialize logger
	logger.InitLogger()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Create the messaging service with confluent-kafka-go
	messagingService := service.NewConfluentMessagingService(cfg)

	// Create the handler with the service
	msgHandler := handler.NewMessagingHandler(cfg, messagingService)

	// Build container
	container := &Container{
		Config:           cfg,
		MessagingHandler: msgHandler,
		MessagingService: messagingService,
	}

	return container, nil
}

// Close properly shuts down all resources
func (c *Container) Close() error {
	// Close messaging service (which will close producers)
	if closer, ok := c.MessagingService.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			logger.LogErrorEvent("", "container_close", "", "error",
				"Failed to close messaging service")
			return err
		}
	}

	return nil
}
