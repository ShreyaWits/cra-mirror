package link

import (
	"fmt"
	"log"
	repository "protected_link/internal/modules/cassandra/repository"
	"protected_link/internal/modules/link_generation/apis/handlers"
	"protected_link/internal/modules/link_generation/repositories"
	"protected_link/internal/modules/link_generation/services"
	database "protected_link/pkg/redis"
)

type Module struct {
	Handler *handlers.GenerateLinkHandler
}

// InitializeModule creates and initializes the link generation module
func InitializeModule(rdb *database.RedisConfig, cassandra repository.ICassandraRepository) (*Module, error) {
	repo, err := repositories.NewGeneratedRepository(rdb, cassandra)
	if err != nil {
		log.Printf("❌ Failed to initialize repository: %v", err)
		return nil, fmt.Errorf("failed to initialize repository: %w", err)
	}

	service := services.NewGenerateLinkService(repo, rdb, cassandra)
	handler := handlers.NewGenerateLinkHandler(service)

	return &Module{
		Handler: handler,
	}, nil
}
