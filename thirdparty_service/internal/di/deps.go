package di

import (
	"log"

	"thirdparty_service/internal/repositories"
	"thirdparty_service/internal/services"

	"go.uber.org/dig"
)

var (
	Container = dig.New()
)

func initializeDeps() error {

	userRepository := repositories.NewPostgresUserRepository()
	if err := Container.Provide(func() repositories.UserRepository {
		return userRepository
	}); err != nil {
		return err
	}

	userService := services.NewUserService(userRepository)
	if err := Container.Provide(func() *services.UserService {
		return userService
	}); err != nil {
		return err
	}

	// ✅ Redis-based WebhookRepository
	redisRepo := repositories.NewRedisRepo("localhost:6379")
	if err := Container.Provide(func() repositories.WebhookRepository {
		return redisRepo
	}); err != nil {
		return err
	}

	// ✅ WebhookService that uses WebhookRepository
	if err := Container.Provide(func(repo repositories.WebhookRepository) services.WebhookService {
		return services.NewWebhookService(repo)
	}); err != nil {
		return err
	}

	return nil
}

func init() {
	if err := initializeDeps(); err != nil {
		log.Fatalf("error while initializing app dependencies: %v\n", err)
	}

	return
}
