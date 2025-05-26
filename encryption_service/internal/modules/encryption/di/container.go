package di

import (
	"context"
	"fmt"
	"time"

	"encryption_microservice/internal/config"
	"encryption_microservice/internal/modules/encryption/api/handlers"
	cacheclient "encryption_microservice/internal/modules/encryption/client/cache_client"
	configclient "encryption_microservice/internal/modules/encryption/client/config_client"
	cachemanager "encryption_microservice/internal/modules/encryption/services/cache_manager"
	service "encryption_microservice/internal/modules/encryption/services/config_manager"
	encryptionengine "encryption_microservice/internal/modules/encryption/services/encryption_engine"
	keymanager "encryption_microservice/internal/modules/encryption/services/key_manager"
	"encryption_microservice/internal/modules/encryption/services/user"
	usecases "encryption_microservice/internal/modules/encryption/usecases/encryption"
	"encryption_microservice/pkg/crypto"
	httpclient "encryption_microservice/pkg/http"
	"encryption_microservice/pkg/kms"
	"encryption_microservice/pkg/observability"
)

// Container holds all the dependencies for the application
type Container struct {
	// Config
	Config *config.Config
	Env    *config.EnvConfig

	// Key Manager
	KeyManager keymanager.KeyManager

	// Encryption Engine
	EncryptionEngine encryptionengine.EncryptionEngine

	// Use Cases
	EncryptionUseCase usecases.EncryptionUseCase

	// Handlers
	EncryptionHandler handlers.EncryptionHandlerImpl
	ConfigHandler     *handlers.ConfigHandler

	// HttpClient
	HttpClient httpclient.HTTPClient

	// Cache Client and Manager
	CacheClient  cacheclient.RedisClient
	CacheManager *cachemanager.CacheManager

	// Config Manager
	ConfigManager *service.ConfigManagerService

	// Observability
	Obs *observability.ObservabilityStack

	// OpenTelemetry shutdown function
	OtelShutdown func(context.Context) error
}

// NewContainer creates a new dependency injection container
func NewContainer() (*Container, error) {
	ctx := context.Background()
	container := &Container{}

	// Load environment configuration
	env, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load environment configuration: %w", err)
	}
	container.Env = env

	// Create observability stack
	obs := observability.NewObservabilityStack(env)
	container.Obs = obs

	// Create HTTP client with timeout
	httpClient, err := httpclient.New(5 * time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}
	container.HttpClient = httpClient

	// Create config client
	configClient, err := configclient.NewConfigClient(httpClient, env)
	if err != nil {
		return nil, fmt.Errorf("failed to create config client: %w", err)
	}

	// Create cache client
	cacheClient, err := cacheclient.NewRedisClient(env.CacheUrl)
	if err != nil {
		container.Obs.LoggerService.Warn(ctx, "Failed to create cache client, continuing without caching: %v", err)
	} else {
		container.CacheClient = cacheClient

		// Create cache manager
		cacheManager, err := cachemanager.NewCacheManager(cacheClient, env, obs)
		if err != nil {
			container.Obs.LoggerService.Warn(ctx, "Failed to create cache manager: %v", err)
		} else {
			container.CacheManager = cacheManager
		}
	}

	configManager, err := service.NewConfigManager(configClient, cacheClient, obs, env)
	if err != nil {
		container.Obs.LoggerService.Warn(ctx, "Failed to create config manager: %v", err)
	} else {
		container.ConfigManager = configManager
	}

	cfg, err := configManager.GetFromApiConfiguration(ctx)
	if err != nil {
		container.Obs.LoggerService.Warn(ctx, "Failed to get config from config manager: %v", err)
	}
	container.Config = cfg

	// Setup OpenTelemetry if we have a configuration
	if cfg.OtelCollectorGrpcEndpoint != "" {
		shutdownOtel, err := observability.SetupOTelSDK(ctx, env, cfg)
		if err != nil {
			container.Obs.LoggerService.Warn(ctx, "Failed to setup OpenTelemetry SDK: %v", err)
			// Create no-op shutdown function to avoid nil checks later
			shutdownOtel = func(context.Context) error { return nil }
		}
		container.OtelShutdown = shutdownOtel
	} else {
		// Create no-op shutdown function
		container.OtelShutdown = func(context.Context) error { return nil }
	}

	// Initialize KMS service
	// Create a config.Config from the EncryptionConfig for backwards compatibility
	legacyConfig := &config.Config{

		VaultAddr:                 cfg.VaultAddr,
		VaultToken:                cfg.VaultToken,
		VaultPath:                 cfg.VaultPath,
		OtelCollectorGrpcEndpoint: cfg.OtelCollectorGrpcEndpoint,
	}
	kmsService := kms.NewHashiCorpKMS(legacyConfig)

	// Initialize key manager
	container.KeyManager = keymanager.NewKeyManager(kmsService, obs)

	// Initialize Crypto
	crypto := crypto.NewEncryption()

	// Initialize user service
	userService := user.NewMockUserService(container.HttpClient, cfg.UserServiceURL)

	// Initialize encryption engine
	container.EncryptionEngine = encryptionengine.NewEncryptionEngine(container.KeyManager, crypto, obs)

	// Initialize use cases with observability
	container.EncryptionUseCase = usecases.NewEncryptionUseCaseWithObs(container.KeyManager, container.EncryptionEngine, obs)

	// Initialize handlers
	container.EncryptionHandler = *handlers.NewEncryptionHandler(container.EncryptionUseCase, userService, obs)

	// Initialize config handler if cache manager is available
	if container.CacheManager != nil {
		configHandler, err := handlers.NewConfigHandler(container.CacheManager, env, obs)
		if err != nil {
			container.Obs.LoggerService.Warn(ctx, "Failed to create config handler: %v", err)
		} else {
			container.ConfigHandler = configHandler
			// Set the reinitialization callback
			container.ConfigHandler.SetReinitCallback(container.Reinitialize)
		}
	}

	return container, nil
}

// Reinitialize reinitializes services with the new configuration
func (c *Container) Reinitialize(newConfig *config.Config) error {
	ctx := context.Background()

	fmt.Println("Reinitializing with new config:", newConfig)
	// Validate and set the config
	if err := config.SetConfig(newConfig); err != nil {
		return fmt.Errorf("failed to validate new configuration: %w", err)
	}

	c.Config = newConfig
	c.Obs.LoggerService.Info(ctx, "Configuration updated successfully")

	// Create a legacy config from the new EncryptionConfig for backward compatibility
	legacyConfig := newConfig

	// Reinitialize the Hashicorp KMS service with the updated configuration
	kmsService := kms.NewHashiCorpKMS(legacyConfig)

	// Reinitialize the key manager with the new KMS service
	c.KeyManager = keymanager.NewKeyManager(kmsService, c.Obs)

	// Reinitialize the encryption engine with the new key manager
	crypto := crypto.NewEncryption()
	c.EncryptionEngine = encryptionengine.NewEncryptionEngine(c.KeyManager, crypto, c.Obs)

	// Reinitialize the encryption use case with the new components
	c.EncryptionUseCase = usecases.NewEncryptionUseCaseWithObs(c.KeyManager, c.EncryptionEngine, c.Obs)

	c.Obs.LoggerService.Info(ctx, "Hashicorp Vault KMS and dependent services reinitialized successfully")

	return nil
}

// Close cleans up resources used by the container
func (c *Container) Close(ctx context.Context) error {
	var err error

	// Close OpenTelemetry
	if c.OtelShutdown != nil {
		if shutdownErr := c.OtelShutdown(ctx); shutdownErr != nil {
			err = fmt.Errorf("error shutting down OpenTelemetry: %w", shutdownErr)
		}
	}

	// Close cache client if it exists
	if c.CacheClient != nil {
		if closeErr := c.CacheClient.Close(); closeErr != nil {
			err = fmt.Errorf("error closing cache client: %w; %v", closeErr, err)
		}
	}

	return err
}

// // SetupRoutes sets up all the routes for the application
// func (c *Container) SetupRoutes(app *fiber.App) {

// 	routes.SetupAPIRoutes(app, c.EncryptionHandler)
// }
