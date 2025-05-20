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

	// Observability
	Obs *observability.ObservabilityStack

	// OpenTelemetry shutdown function
	OtelShutdown func(context.Context) error
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
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Setup OpenTelemetry
	shutdownOtel, err := observability.SetupOTelSDK(ctx, env)
	if err != nil {
		return nil, fmt.Errorf("failed to setup OpenTelemetry SDK: %w", err)
	}
	if shutdownOtel == nil {
		// Create no-op shutdown function to avoid nil checks later
		shutdownOtel = func(context.Context) error { return nil }
	}

	// Create observability stack
	obs := &observability.ObservabilityStack{
		TracerService:  tracer.NewTracer(env.ServiceName, true),
		MetricsService: metrics.NewMetricsService(env.ServiceName, true),
		LoggerService:  logger.NewLogger(env.ServiceName, true),
	}

	// Validate observability components
	if obs.TracerService == nil || obs.MetricsService == nil || obs.LoggerService == nil {
		return nil, fmt.Errorf("failed to initialize observability components")
	}

	// Create HTTP client with timeout
	httpClient, err := httpclient.New(5 * time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Create config client
	configClient, err := client.NewConfigClient(httpClient, env)
	if err != nil {
		return nil, fmt.Errorf("failed to create config client: %w", err)
	}

	// Create cache client
	cacheClient, err := cacheclient.NewRedisClient(env.CacheUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache client: %w", err)
	}
	if cacheClient == nil {
		return nil, fmt.Errorf("cache client is nil despite no error")
	}

	// Create config manager service
	configManagerService, err := service.NewConfigManager(configClient, cacheClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create config manager service: %w", err)
	}

	// Fetch configuration
	cfg, err := configManagerService.GetFromApiConfiguration()
	if err != nil {
		return nil, fmt.Errorf("not able to fetch config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("configuration is nil")
	}

	// Create Kafka factory
	factory, err := confluent.NewFactory(obs)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka factory: %w", err)
	}

	// Create messaging service
	messagingService, err := service.NewConfluentMessagingService(cfg, factory, obs)
	if err != nil {
		return nil, fmt.Errorf("failed to create messaging service: %w", err)
	}
	if messagingService == nil {
		return nil, fmt.Errorf("messaging service is nil despite no error")
	}

	// Create messaging handler
	msgHandler, err := handler.NewMessagingHandler(cfg, messagingService, obs)
	if err != nil {
		return nil, fmt.Errorf("failed to create messaging handler: %w", err)
	}

	// Create config handler with environment
	configHandler, err := handler.NewConfigHandler(configManagerService, env)
	if err != nil {
		return nil, fmt.Errorf("failed to create config handler: %w", err)
	}

	// Build container
	container := &Container{
		Config:           cfg,
		Env:              env,
		MessagingHandler: msgHandler,
		MessagingService: messagingService,
		ConfigHandler:    configHandler,
		Obs:              obs,
		OtelShutdown:     shutdownOtel,
	}

	// Set the reinitialize callback
	configHandler.SetReinitCallback(container.Reinitialize)

	// Validate the container before returning
	if err := container.ValidateContainer(); err != nil {
		return nil, fmt.Errorf("container validation failed: %w", err)
	}

	return container, nil
}

// Reinitialize reinitializes services with the new configuration
func (c *Container) Reinitialize(newConfig *config.Config) error {
	ctx := context.Background()

	// Compare configs to see if we need to reinitialize
	if !configRequiresReinitialization(c.Config, newConfig) {
		// Validate and set the config, even if no reinitialization is needed
		if err := config.SetConfig(newConfig); err != nil {
			return fmt.Errorf("failed to validate new configuration: %w", err)
		}

		c.Config = newConfig // Still update the config reference
		c.Obs.LoggerService.Info(ctx, "Configuration updated, but no critical changes detected requiring reinitialization")
		return nil
	}

	c.Obs.LoggerService.Info(ctx, "Critical configuration changes detected, reinitializing services", map[string]interface{}{
		"service":      c.Env.ServiceName,
		"kafkaBrokers": newConfig.KafkaBrokers,
	})

	// Validate the configuration before proceeding
	if err := config.SetConfig(newConfig); err != nil {
		return fmt.Errorf("failed to validate new configuration: %w", err)
	}

	// Only close messaging service without shutting down OTel
	if c.MessagingService != nil {
		if closer, ok := c.MessagingService.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				c.Obs.LoggerService.Error(ctx, "Error closing messaging service: %v", err)
				// Continue anyway, attempt to recreate
			} else {
				c.Obs.LoggerService.Info(ctx, "Successfully closed messaging service")
			}
		}
	}

	// Keep using the existing OTel and observability stack
	// Create a new factory with existing observability stack
	factory, err := confluent.NewFactory(c.Obs)
	if err != nil {
		return fmt.Errorf("failed to create Kafka factory during reinitialization: %w", err)
	}

	// Create a new messaging service
	messagingService, err := service.NewConfluentMessagingService(newConfig, factory, c.Obs)
	if err != nil {
		return fmt.Errorf("failed to reinitialize messaging service: %w", err)
	}

	// Create a new handler with the new service
	msgHandler, err := handler.NewMessagingHandler(newConfig, messagingService, c.Obs)
	if err != nil {
		return fmt.Errorf("failed to create messaging handler during reinitialization: %w", err)
	}

	// Update container references
	c.Config = newConfig
	c.MessagingService = messagingService
	c.MessagingHandler = msgHandler

	// Validate container after reinitializing
	if err := c.ValidateContainer(); err != nil {
		return fmt.Errorf("container validation after reinitialization failed: %w", err)
	}

	c.Obs.LoggerService.Info(ctx, "Services successfully reinitialized")

	return nil
}

// configRequiresReinitialization determines if the configuration changes require a service restart
func configRequiresReinitialization(oldConfig, newConfig *config.Config) bool {
	// Don't reinitialize if configs are identical
	if oldConfig == newConfig {
		return false
	}

	// Check for changes to critical configuration parameters
	if oldConfig == nil || newConfig == nil {
		return true
	}

	// Check Kafka broker changes - the most critical configuration
	if !slicesEqual(oldConfig.KafkaBrokers, newConfig.KafkaBrokers) {
		return true
	}

	// Check other critical Kafka configuration parameters
	if oldConfig.KafkaNumPartitions != newConfig.KafkaNumPartitions ||
		oldConfig.KafkaReplicationFactor != newConfig.KafkaReplicationFactor ||
		oldConfig.KafkaRequireActiveListener != newConfig.KafkaRequireActiveListener {
		return true
	}

	// Check transaction and reliability settings
	if oldConfig.KafkaBatchTimeoutMs != newConfig.KafkaBatchTimeoutMs ||
		oldConfig.KafkaMaxAttempts != newConfig.KafkaMaxAttempts {
		return true
	}

	// Other parameters might not require full reinitialization
	return false
}

// slicesEqual compares two string slices for equality
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// Close properly shuts down all resources
func (c *Container) Close(ctx context.Context) error {
	var err error

	// Close messaging service (which will close producers)
	if closer, ok := c.MessagingService.(interface{ Close() error }); ok {
		if closeErr := closer.Close(); closeErr != nil {
			c.Obs.LoggerService.Error(ctx, "Failed to close messaging service %v", closeErr)
			err = closeErr
		}
	}

	// Shutdown OpenTelemetry if the shutdown function is available
	if c.OtelShutdown != nil {
		if otelErr := c.OtelShutdown(ctx); otelErr != nil {
			c.Obs.LoggerService.Error(ctx, "Failed to shutdown OpenTelemetry %v", otelErr)
			// Join the errors
			if err != nil {
				err = fmt.Errorf("multiple shutdown errors: %w and OpenTelemetry: %v", err, otelErr)
			} else {
				err = otelErr
			}
		}
	}

	return err
}

// ValidateContainer performs a comprehensive validation of all container components
func (c *Container) ValidateContainer() error {
	if c == nil {
		return fmt.Errorf("container is nil")
	}

	if err := validateConfig(c.Config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	if err := validateObservability(c.Obs); err != nil {
		return fmt.Errorf("observability validation failed: %w", err)
	}

	if err := validateMessagingComponents(c.MessagingService, c.MessagingHandler); err != nil {
		return fmt.Errorf("messaging components validation failed: %w", err)
	}

	if c.ConfigHandler == nil {
		return fmt.Errorf("config handler is nil")
	}

	// Validate ConfigHandler has valid ConfigManager service
	if err := validateConfigHandler(c.ConfigHandler); err != nil {
		return fmt.Errorf("config handler validation failed: %w", err)
	}

	return nil
}

// validateConfig checks if the configuration is valid
func validateConfig(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	// Use the validator to validate the config structure
	if err := config.SetConfig(cfg); err != nil {
		return err
	}

	return nil
}

// validateObservability checks if the observability stack is properly configured
func validateObservability(obs *observability.ObservabilityStack) error {
	if obs == nil {
		return fmt.Errorf("observability stack is nil")
	}

	if obs.TracerService == nil {
		return fmt.Errorf("tracer service is nil")
	}

	if obs.MetricsService == nil {
		return fmt.Errorf("metrics service is nil")
	}

	if obs.LoggerService == nil {
		return fmt.Errorf("logger service is nil")
	}

	return nil
}

// validateMessagingComponents checks if the messaging components are properly configured
func validateMessagingComponents(messagingService service.MessagingService, messagingHandler *handler.MessagingHandler) error {
	if messagingService == nil {
		return fmt.Errorf("messaging service is nil")
	}

	if messagingHandler == nil {
		return fmt.Errorf("messaging handler is nil")
	}

	return nil
}

// validateConfigHandler checks if the config handler is properly configured
func validateConfigHandler(configHandler *handler.ConfigHandler) error {
	if configHandler == nil {
		return fmt.Errorf("config handler is nil")
	}

	// Note: The ConfigManager service is validated indirectly here.
	// When the ConfigHandler was created, it received the ConfigManagerService as a dependency.
	// If the ConfigManagerService was invalid, the NewConfigHandler function would have
	// returned an error at that point, and we wouldn't have a valid ConfigHandler now.

	return nil
}
