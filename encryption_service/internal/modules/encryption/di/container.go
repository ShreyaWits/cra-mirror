package di

import (
	"encryption_microservice/internal/config"

	"encryption_microservice/internal/modules/encryption/api/handlers"
	encryptionengine "encryption_microservice/internal/modules/encryption/services/encryption_engine"
	keymanager "encryption_microservice/internal/modules/encryption/services/key_manager"
	"encryption_microservice/internal/modules/encryption/services/user"
	usecases "encryption_microservice/internal/modules/encryption/usecases/encryption"
	"encryption_microservice/pkg/crypto"
	"encryption_microservice/pkg/http"
	"encryption_microservice/pkg/kms"
	"encryption_microservice/pkg/logger"
)

// Container holds all the dependencies for the application
type Container struct {
	// Config
	Config *config.Config

	// Key Manager
	KeyManager keymanager.KeyManager

	// Encryption Engine
	EncryptionEngine encryptionengine.EncryptionEngine

	// Use Cases
	EncryptionUseCase usecases.EncryptionUseCase

	// Handlers
	EncryptionHandler handlers.EncryptionHandlerImpl

	// HttpClient
	HttpClient http.HttpClient
}

// NewContainer creates a new dependency injection container
func NewContainer() (*Container, error) {
	container := &Container{}
	cfg, err := config.LoadConfig()
	// Load configuration
	if err != nil {
		return nil, err
	}
	container.Config = cfg

	logger.InitLogger()

	// Initialize KMS service
	kmsService := kms.NewHashiCorpKMS(cfg)

	// Initialize key manager
	container.KeyManager = keymanager.NewKeyManager(kmsService)

	// Initialize Crypto
	crypto := crypto.NewEncryption()

	userService := user.NewMockUserService(container.HttpClient, cfg.UserServiceURL)

	// Initialize encryption engine
	container.EncryptionEngine = encryptionengine.NewEncryptionEngine(container.KeyManager, crypto)

	// Initialize use cases
	container.EncryptionUseCase = usecases.NewEncryptionUseCase(container.KeyManager, container.EncryptionEngine)

	// Initialize handlers
	container.EncryptionHandler = *handlers.NewEncryptionHandler(container.EncryptionUseCase, userService)

	return container, nil
}

// // SetupRoutes sets up all the routes for the application
// func (c *Container) SetupRoutes(app *fiber.App) {

// 	routes.SetupAPIRoutes(app, c.EncryptionHandler)
// }
