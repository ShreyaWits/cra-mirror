package di

import (
	"Document-Processing/internal/config"
	"Document-Processing/internal/grpc"
	"Document-Processing/internal/handlers"
	"Document-Processing/internal/services"
)

type Container struct {
	Config          *config.Config
	Server          *grpc.Server
	DocumentService services.DocumentService
	DocumentHandler *handlers.DocumentHandler
	HealthHandler   *handlers.HealthHandler
}

func NewContainer() *Container {
	cfg := config.NewConfig()

	// Initialize services
	documentService := services.NewDocumentService()

	// Initialize handlers
	documentHandler := handlers.NewDocumentHandler(documentService)
	healthHandler := handlers.NewHealthHandler()

	// Initialize server
	server := grpc.NewServer(cfg.ServerPort)
	server.RegisterServices(documentHandler, healthHandler)

	return &Container{
		Config:          cfg,
		Server:          server,
		DocumentService: documentService,
		DocumentHandler: documentHandler,
		HealthHandler:   healthHandler,
	}
}
