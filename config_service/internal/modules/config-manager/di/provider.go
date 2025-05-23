package di

import (
	"fmt"
	"nps-config-service/internal/configs"
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"
	"nps-config-service/internal/modules/config-manager/repositories"
	"nps-config-service/internal/modules/config-manager/services"
	etcdDB "nps-config-service/pkg/etcd"
	"nps-config-service/pkg/observability"
	workflows "nps-config-service/pkg/temporal"

	temporalClient "go.temporal.io/sdk/client"
)

// Container holds all dependencies for the application
type Container struct {
	// Infrastructure
	EtcdClient         *etcdDB.EtcdClientImpl
	TemporalClient     temporalClient.Client
	ObservabilityStack *observability.ObservabilityStack

	// Repositories
	ConfigRepo repositories.IConfigRepo

	// Services
	AdminService   *services.AdminService
	WebhookService *services.WebhookService
	ConfigService  *services.ConfigService

	// Handlers
	AdminHandler   *handler.AdminHandler
	ConfigHandler  *handler.ConfigHandler
	WebhookHandler *handler.WebhookHandler
}

// ContainerOption is a function that configures the Container
type ContainerOption func(*Container) error

// NewContainer creates a new Container with the given options
func NewContainer(opts ...ContainerOption) (*Container, error) {
	container := &Container{}

	// Apply all options
	for _, opt := range opts {
		if err := opt(container); err != nil {
			return nil, fmt.Errorf("failed to apply container option: %w", err)
		}
	}

	return container, nil
}

// WithInfrastructure initializes infrastructure dependencies
func WithInfrastructure() ContainerOption {
	return func(c *Container) error {
		// Initialize observability
		observabilityStack := observability.NewObservabilityStack("config-service")
		c.ObservabilityStack = observabilityStack

		// Initialize etcd client
		client, err := etcdDB.InitEtcdDB(configs.AppConfig.EtcdEndpoint)
		if err != nil {
			return fmt.Errorf("failed to initialize etcd client: %w", err)
		}
		c.EtcdClient = etcdDB.NewEtcdClientImpl(client)

		// Initialize temporal client
		temporalClient, err := workflows.InitTemporal(configs.AppConfig.TemporalEndpoint)
		if err != nil {
			return fmt.Errorf("failed to initialize temporal client: %w", err)
		}
		c.TemporalClient = *temporalClient

		return nil
	}
}

// WithRepositories initializes repository layer
func WithRepositories() ContainerOption {
	return func(c *Container) error {
		if c.ObservabilityStack == nil {
			return fmt.Errorf("observability stack not initialized")
		}
		if c.EtcdClient == nil {
			return fmt.Errorf("etcd client not initialized")
		}

		c.ConfigRepo = repositories.NewConfigRepository(c.EtcdClient, c.ObservabilityStack)
		return nil
	}
}

// WithServices initializes service layer
func WithServices() ContainerOption {
	return func(c *Container) error {
		if c.ConfigRepo == nil {
			return fmt.Errorf("config repository not initialized")
		}
		if c.ObservabilityStack == nil {
			return fmt.Errorf("observability stack not initialized")
		}
		if c.TemporalClient == nil {
			return fmt.Errorf("temporal client not initialized")
		}

		// Initialize services
		adminService := services.NewAdminService(c.ConfigRepo, c.ObservabilityStack)
		webhookService := services.NewWebhookService(c.ConfigRepo, c.ObservabilityStack)
		configService := services.NewConfigService(c.ConfigRepo, webhookService, c.TemporalClient, c.ObservabilityStack)

		// Type assertions
		var ok bool
		c.AdminService, ok = adminService.(*services.AdminService)
		if !ok {
			return fmt.Errorf("failed to type assert admin service")
		}

		c.WebhookService, ok = webhookService.(*services.WebhookService)
		if !ok {
			return fmt.Errorf("failed to type assert webhook service")
		}

		c.ConfigService, ok = configService.(*services.ConfigService)
		if !ok {
			return fmt.Errorf("failed to type assert config service")
		}

		return nil
	}
}

// WithHandlers initializes handler layer
func WithHandlers() ContainerOption {
	return func(c *Container) error {
		if c.AdminService == nil || c.ConfigService == nil || c.WebhookService == nil {
			return fmt.Errorf("services not initialized")
		}
		if c.ObservabilityStack == nil {
			return fmt.Errorf("observability stack not initialized")
		}

		c.AdminHandler = handler.NewAdminHandler(c.AdminService, c.ObservabilityStack, nil)
		c.ConfigHandler = handler.NewConfigHandler(c.ConfigService, c.ObservabilityStack)
		c.WebhookHandler = handler.NewWebhookHandler(c.WebhookService, c.ObservabilityStack)

		return nil
	}
}

// InitContainer initializes a new container with all dependencies
func InitContainer() (*Container, error) {
	return NewContainer(
		WithInfrastructure(),
		WithRepositories(),
		WithServices(),
		WithHandlers(),
	)
}

// Close gracefully shuts down the container
func (c *Container) Close() error {
	// Add cleanup logic here if needed
	// For example, closing database connections, etc.
	return nil
}

// GetHandlers returns all handlers from the container
func (c *Container) GetHandlers() (*handler.AdminHandler, *handler.ConfigHandler, *handler.WebhookHandler, error) {
	if c.AdminHandler == nil || c.ConfigHandler == nil || c.WebhookHandler == nil {
		return nil, nil, nil, fmt.Errorf("handlers not initialized")
	}
	return c.AdminHandler, c.ConfigHandler, c.WebhookHandler, nil
}
