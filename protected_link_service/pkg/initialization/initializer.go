package initialization

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"protected_link/internal/common/utils"
	configEnv "protected_link/internal/configs"
	auth "protected_link/internal/modules/authentication"
	"protected_link/internal/modules/cassandra/infrastructure"
	link "protected_link/internal/modules/link_generation"
	"protected_link/internal/server"
	"protected_link/pkg/cassandra"
	database "protected_link/pkg/redis"
	"protected_link/pkg/validation"
)

// AppConfig holds all the application configuration and dependencies
type AppConfig struct {
	Config         *configEnv.Config
	Cassandra      *cassandra.CassandraConfig
	Redis          *database.RedisConfig
	AuthModule     *auth.Module
	LinkModule     *link.Module
	GRPCServer     *server.GRPCServer
	ShutdownSignal chan os.Signal
}

// NewAppConfig creates a new application configuration
func NewAppConfig() (*AppConfig, error) {
	cfg, err := configEnv.LoadConfig()
	if err != nil {
		return nil, err
	}

	return &AppConfig{
		Config:         cfg,
		ShutdownSignal: make(chan os.Signal, 1),
	}, nil
}

// InitializeDatabaseConnections sets up database connections with retry logic
func (a *AppConfig) InitializeDatabaseConnections() error {
	// Initialize Cassandra
	cassandraConfig, err := cassandra.NewCassandraConfig(a.Config)
	if err != nil {
		return err
	}
	a.Cassandra = cassandraConfig

	// Initialize Redis
	redisConfig, err := database.ConnectRedis(a.Config)
	if err != nil {
		return err
	}
	a.Redis = redisConfig

	return nil
}

// InitializeModules sets up all application modules
func (a *AppConfig) InitializeModules() error {
	validation.InitValidator()

	cassandraClient := infrastructure.NewCassandraClient(a.Cassandra)
	cassandraRepo := infrastructure.NewCassandraRepository(cassandraClient)

	a.AuthModule = auth.InitializeModule(a.Redis, cassandraRepo)

	linkModule, err := link.InitializeModule(a.Redis, cassandraRepo)
	if err != nil {
		return fmt.Errorf("failed to initialize link module: %w", err)
	}
	a.LinkModule = linkModule

	return nil
}

// InitializeServer sets up the gRPC server
func (a *AppConfig) InitializeServer() {
	a.GRPCServer = server.NewGRPCServer(a.AuthModule.Handler, a.LinkModule.Handler)
}

// SetupShutdownHandler configures graceful shutdown
func (a *AppConfig) SetupShutdownHandler() {
	signal.Notify(a.ShutdownSignal, syscall.SIGINT, syscall.SIGTERM)
}

// StartServer starts the gRPC server in a goroutine
func (a *AppConfig) StartServer() error {
	go func() {
		if err := a.GRPCServer.Start(a.Config.GRPCPort); err != nil {
			log.Fatalf("❌ Failed to start gRPC server: %v", err)
		}
	}()
	return nil
}

// WaitForShutdown waits for shutdown signal and performs cleanup
func (a *AppConfig) WaitForShutdown() {
	<-a.ShutdownSignal
	log.Println("🛑 Shutting down server...")

	// Cleanup resources
	if a.Cassandra != nil {
		a.Cassandra.Close()
	}
	if a.Redis != nil {
		a.Redis.Close()
	}
}

// InitializeApp performs all initialization steps
func InitializeApp() (*AppConfig, error) {
	app, err := NewAppConfig()
	if err != nil {
		return nil, err
	}

	if err := app.InitializeDatabaseConnections(); err != nil {
		return nil, err
	}

	if err := app.InitializeModules(); err != nil {
		return nil, err
	}

	utils.LoadMessages()

	app.InitializeServer()
	app.SetupShutdownHandler()

	return app, nil
}
