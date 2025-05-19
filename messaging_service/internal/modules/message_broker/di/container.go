package di

import (
	"context"
	"fmt"
	"messaging_service/internal/config"
	"messaging_service/internal/modules/message_broker/api/handler"
	cacheclient "messaging_service/internal/modules/message_broker/client/cache_client"
	client "messaging_service/internal/modules/message_broker/client/config_client"
	"messaging_service/internal/modules/message_broker/service"
	"messaging_service/pkg/confluent"
	httpclient "messaging_service/pkg/http"
	"messaging_service/pkg/logger"
	metrics "messaging_service/pkg/matrics"
	"messaging_service/pkg/observability"
	"messaging_service/pkg/tracer"
	"time"
)

// Container holds all the dependencies for the application
type Container struct {
	// Config
	Config *config.Config
	Env    *config.Env

	// Service handlers
	MessagingHandler *handler.MessagingHandler
	ConfigHandler    *handler.ConfigHandler

	// Services
	MessagingService service.MessagingService

	Obs *observability.ObservabilityStack
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

	ctx := context.Background()
	// Load configuration
	env, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Then setup OpenTelemetry
	observability.SetupOTelSDK(ctx, env)

	obs := &observability.ObservabilityStack{
		TracerService:  tracer.NewTracer(env.ServiceName, true),
		MetricsService: metrics.NewMetricsService(env.ServiceName, true),
		LoggerService:  logger.NewLogger(env.ServiceName, true),
	}

	httpClient := httpclient.New(5 * time.Second)
	configClient := client.NewConfigClient(httpClient, env)
	cacheClient, cacheErr := cacheclient.NewRedisClient(env.CacheUrl)

	if cacheErr != nil {
		return nil, cacheErr
	}
	// Create the config manager service
	configManagerService := service.NewConfigManager(configClient, cacheClient)
	// The service doesn't have a constructor, so we ini

	cfg, err := configManagerService.GetFromApiConfiguration()
	if err != nil {
		return nil, fmt.Errorf("not able to fetch config %v", err)
	}

	factory := confluent.NewFactory(obs)
	// Create the messaging service with confluent-kafka-go
	messagingService, err := service.NewConfluentMessagingService(cfg, factory, obs)
	if err != nil {
		return nil, err
	}

	// Create the handler with the service
	msgHandler := handler.NewMessagingHandler(cfg, messagingService, obs)

	// Create the config handler using the constructor
	configHandler := handler.NewConfigHandler(configManagerService)

	// Build container
	container := &Container{
		Config:           cfg,
		Env:              env,
		MessagingHandler: msgHandler,
		MessagingService: messagingService,
		ConfigHandler:    configHandler,
		Obs:              obs,
	}

	return container, nil
}

// Close properly shuts down all resources
func (c *Container) Close(ctx context.Context) error {
	// Close messaging service (which will close producers)
	if closer, ok := c.MessagingService.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			c.Obs.LoggerService.Error(ctx, "Failed to close messaging service %v", err)

			return err
		}
	}

	return nil
}
