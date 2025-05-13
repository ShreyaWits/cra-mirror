package app

import (
	"thirdparty_service/internal/config"
	user_repo "thirdparty_service/internal/modules/user/repository"
	user_service "thirdparty_service/internal/modules/user/service"
	redis_repo "thirdparty_service/internal/modules/webhooks/repository"
	webhook_service "thirdparty_service/internal/modules/webhooks/service"

	"go.uber.org/dig"
)

var Container *dig.Container

func InitDependencyInjection() error {
	Container = dig.New()

	userRepository := user_repo.NewPostgresUserRepository()
	if err := Container.Provide(func() user_repo.UserRepository {
		return userRepository
	}); err != nil {
		return err
	}

	userService := user_service.NewUserService(userRepository)
	if err := Container.Provide(func() *user_service.UserService {
		return userService
	}); err != nil {
		return err
	}

	redisRepo := redis_repo.NewRedisRepo(config.AppConfig.GetRedisAddress())
	if err := Container.Provide(func() redis_repo.WebhookRepository {
		return redisRepo
	}); err != nil {
		return err
	}

	// ✅ WebhookService that uses WebhookRepository
	if err := Container.Provide(func(repo redis_repo.WebhookRepository) webhook_service.WebhookService {
		return webhook_service.NewWebhookService(repo)
	}); err != nil {
		return err
	}
	return nil
}

func GetUserRepository() user_repo.UserRepository {
	var repo user_repo.UserRepository
	if err := Container.Invoke(func(r user_repo.UserRepository) {
		repo = r
	}); err != nil {
		panic(err)
	}
	return repo
}

func GetWebhookRepository() redis_repo.WebhookRepository {
	var repo redis_repo.WebhookRepository
	if err := Container.Invoke(func(r redis_repo.WebhookRepository) {
		repo = r
	}); err != nil {
		panic(err)
	}
	return repo
}

func GetUserService() *user_service.UserService {
	var service *user_service.UserService
	if err := Container.Invoke(func(s *user_service.UserService) {
		service = s
	}); err != nil {
		panic(err)
	}
	return service
}

func GetWebhookService() webhook_service.WebhookService {
	var service webhook_service.WebhookService
	if err := Container.Invoke(func(s webhook_service.WebhookService) {
		service = s
	}); err != nil {
		panic(err)
	}
	return service
}

func GetRedisService() *redis_repo.RedisRepo {
	var service *redis_repo.RedisRepo
	if err := Container.Invoke(func(s *redis_repo.RedisRepo) {
		service = s
	}); err != nil {
		panic(err)
	}
	return service
}
