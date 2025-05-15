package app

import (
	"log/slog"
	"thirdparty_service/internal/config"
	user_repo "thirdparty_service/internal/modules/user/repository"
	user_service "thirdparty_service/internal/modules/user/service"
	redis_repo "thirdparty_service/internal/modules/webhooks/repository"
	webhook_service "thirdparty_service/internal/modules/webhooks/service"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
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

	// Get the global tracer provider
	tracerProvider := otel.GetTracerProvider()
	if tracerProvider == nil {
		// If no tracer provider is set, use no-op provider
		tracerProvider = noop.NewTracerProvider()
		otel.SetTracerProvider(tracerProvider)
	}

	tracer := tracerProvider.Tracer(config.AppConfig.SERVICE_NAME)
	logger := otelslog.NewLogger(config.AppConfig.SERVICE_NAME)
	metricMeter := otel.Meter(config.AppConfig.SERVICE_NAME)

	if err := Container.Provide(func() trace.Tracer {
		return tracer
	}); err != nil {
		return err
	}
	if err := Container.Provide(func() *slog.Logger {
		return logger
	}); err != nil {
		return err
	}
	if err := Container.Provide(func() metric.Meter {
		return metricMeter
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

func GetTracer() trace.Tracer {
	var tracer trace.Tracer
	if err := Container.Invoke(func(t trace.Tracer) {
		tracer = t
	}); err != nil {
		panic(err)
	}
	return tracer
}
func GetLogger() *slog.Logger {
	var logger *slog.Logger
	if err := Container.Invoke(func(l *slog.Logger) {
		logger = l
	}); err != nil {
		panic(err)
	}
	return logger
}
func GetMeter() metric.Meter {
	var meter metric.Meter
	if err := Container.Invoke(func(m metric.Meter) {
		meter = m
	}); err != nil {
		panic(err)
	}
	return meter
}
