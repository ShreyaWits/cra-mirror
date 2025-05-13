package di

import (
	"messaging_service/internal/config"
	"messaging_service/internal/messaging_service/handler"
	"messaging_service/internal/messaging_service/service"
	"messaging_service/pkg/kafka"
	"messaging_service/pkg/logger"
	"time"
)

// Container holds all the dependencies for the application
type Container struct {
	// Config
	Config *config.Config

	// Service handlers
	MessagingHandler *handler.MessagingHandler

	// Services
	MessagingService service.MessagingService

	// Repositories and infrastructure
	KafkaAdmin kafka.KafkaAdmin
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

	// Initialize dependencies with environment-based configuration
	kafkaConfig := kafka.KafkaConfig{
		Brokers:           cfg.KafkaBrokers,
		NumPartitions:     cfg.KafkaNumPartitions,
		ReplicationFactor: cfg.KafkaReplicationFactor,
		BatchSize:         cfg.KafkaBatchSize,
		BatchBytes:        cfg.KafkaBatchBytes,
		BatchTimeout:      time.Duration(cfg.KafkaBatchTimeoutMs) * time.Millisecond,
		CompressionCodec:  cfg.KafkaCompressionCodec,
		MaxAttempts:       cfg.KafkaMaxAttempts,
		RetryBackoffMs:    cfg.KafkaRetryBackoffMs,
		ReadTimeout:       time.Duration(cfg.KafkaReadTimeoutMs) * time.Millisecond,
		WriteTimeout:      time.Duration(cfg.KafkaWriteTimeoutMs) * time.Millisecond,
	}

	admin := kafka.NewAdmin(kafkaConfig)

	// Create the messaging service with the admin
	messagingService := service.NewMessagingService(admin)

	// Create the handler with the service
	msgHandler := handler.NewMessagingHandler(cfg, messagingService)

	// Build container
	container := &Container{
		Config:           cfg,
		MessagingHandler: msgHandler,
		MessagingService: messagingService,
		KafkaAdmin:       admin,
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

	// Close admin client if needed
	if closer, ok := c.KafkaAdmin.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			logger.LogErrorEvent("", "container_close", "", "error",
				"Failed to close Kafka admin client")
			return err
		}
	}

	return nil
}
