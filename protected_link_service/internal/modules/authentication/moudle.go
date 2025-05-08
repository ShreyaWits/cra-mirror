package auth

import (
	"protected_link/internal/modules/authentication/api/handlers"
	authRepository "protected_link/internal/modules/authentication/repositories"
	"protected_link/internal/modules/authentication/services"
	repository "protected_link/internal/modules/cassandra/repository"
	database "protected_link/pkg/redis"
)

type Module struct {
	Handler *handlers.AuthHandler
}

func InitializeModule(rdb *database.RedisConfig, casendra repository.ICassandraRepository) *Module {
	repo := authRepository.NewOTPRepository(rdb, casendra)
	service := services.NewAuthenticationService(repo, casendra)
	handler := handlers.NewAuthHandler(service)

	return &Module{
		Handler: handler,
	}
}
