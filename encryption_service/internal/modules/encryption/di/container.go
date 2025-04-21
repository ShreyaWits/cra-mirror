package di

import (
	"encryption_microservice/internal/app/routes"
	"encryption_microservice/internal/config"
	"log"
	"os"
	"time"

	"encryption_microservice/internal/modules/encryption/api/handlers"
	encryptionengine "encryption_microservice/internal/modules/encryption/services/encryption_engine"
	keymanager "encryption_microservice/internal/modules/encryption/services/key_manager"
	"encryption_microservice/internal/modules/encryption/services/user"
	usecases "encryption_microservice/internal/modules/encryption/usecases/encryption"
	"encryption_microservice/pkg/http"
	"encryption_microservice/pkg/kms"

	"github.com/gofiber/fiber/v2"
	"github.com/hashicorp/vault/api"
)

// Container holds all the dependencies for the application
type Container struct {
	// Config
	Config *config.Config

	// KMS
	KMSClient kms.KmsService

	// Key Manager
	KeyManager keymanager.KeyManager

	// Encryption Engine
	EncryptionEngine encryptionengine.EncryptionEngine

	// Use Cases
	EncryptionUseCase usecases.EncryptionUseCase

	// Handlers
	EncryptionHandler handlers.EncryptionHandler

	// HttpClient
	HttpClient http.HttpClient
}

// NewContainer creates a new dependency injection container
func NewContainer() (*Container, error) {
	container := &Container{}

	// Initialize Vault client
	vaultConfig := api.DefaultConfig()
	vaultConfig.Address = os.Getenv("VAULT_ADDR")
	vaultConfig.MaxRetries = 3
	vaultConfig.Timeout = 10 * time.Second

	vaultClient, err := api.NewClient(vaultConfig)
	if err != nil {
		log.Fatalf("Failed to create Vault client: %v", err)
	}

	// Set Vault token
	vaultClient.SetToken(os.Getenv("VAULT_TOKEN"))

	// Initialize KMS service
	kmsService := kms.NewHashiCorpKMS(vaultClient, os.Getenv("VAULT_PATH"))

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	container.Config = cfg

	// Initialize KMS client
	// kmsClient, err := kms_pkg.NewHashicorpKMS(
	// 	cfg.VaultAddr,
	// 	cfg.VaultToken,
	// 	cfg.VaultPath,
	// )

	// if err != nil {
	// 	return nil, err
	// }

	httpClient := http.NewHttpClient()
	kmsClient := kms.NewTemporaryMockKMS()
	userService := user.NewUserService(httpClient, cfg.UserServiceURL)

	container.KMSClient = kmsClient

	// Initialize key manager
	container.KeyManager = keymanager.NewKeyManager(kmsService)

	// Initialize encryption engine
	container.EncryptionEngine = encryptionengine.NewEncryptionEngine(container.KeyManager, nil)

	// Initialize use cases
	container.EncryptionUseCase = usecases.NewEncryptionUseCase(container.KeyManager, container.EncryptionEngine, userService)

	// Initialize handlers
	container.EncryptionHandler = handlers.NewEncryptionHandler(container.EncryptionUseCase)

	return container, nil
}

// SetupRoutes sets up all the routes for the application
func (c *Container) SetupRoutes(app *fiber.App) {
	routes.SetupAPIRoutes(app, c.EncryptionHandler)
}
