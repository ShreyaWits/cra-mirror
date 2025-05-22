//go:build !test
// build +test
package di

import (
	"Document-Processing/internal/config"
	"Document-Processing/internal/grpc"
	"Document-Processing/internal/handlers"
	"Document-Processing/internal/repository"
	"Document-Processing/internal/services"
	"Document-Processing/pkg/yugabytedb"
	"log"
)

type Container struct {
	Config          *config.Config
	Server          *grpc.Server
	DocumentService *services.DocumentService
	HealthHandler   *handlers.HealthHandler
}

func NewContainer(cfg *config.Config) (*Container, error) {
	// Initialize MinIO repository
	minioRepo, err := repository.NewMinioRepository(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioBucketName,
	)
	if err != nil {
		return nil, err
	}

	db, err := yugabytedb.ConnectDB(cfg.YugabyteHost, cfg.YugabyteName, cfg.YugabytePassWord, cfg.YugabytePort, cfg.YugabyteUser)
	if err != nil {
		return nil, err
	}

	// DB yugabyte connection
	yugabyteRepo := repository.NewDocumentDataRepository(db)

	// Initialize Gemini service
	geminiService, err := services.NewGeminiService(cfg.GeminiAPIKey, minioRepo)
	if err != nil {
		return nil, err
	}

	// Initialize Llama service
	llamaService := services.NewLlamaService()

	// Initialize services
	documentService := services.NewDocumentService(geminiService, llamaService, minioRepo, yugabyteRepo)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()

	// Initialize server
	server := grpc.NewServer(cfg.ServerPort)
	server.RegisterServices(documentService, healthHandler)

	return &Container{
		Config:          cfg,
		Server:          server,
		DocumentService: documentService,
		HealthHandler:   healthHandler,
	}, nil
}

// Close closes all resources in the container
func (c *Container) Close() error {
	if c.DocumentService != nil && c.DocumentService.GeminiService() != nil {
		if err := c.DocumentService.GeminiService().Close(); err != nil {
			log.Printf("Error closing Gemini service: %v", err)
		}
	}
	return nil
}
