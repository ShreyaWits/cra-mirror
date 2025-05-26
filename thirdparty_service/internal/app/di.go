package app

import (
	"log"
	"log/slog"
	"thirdparty_service/internal/config"
	"thirdparty_service/internal/modules/config/handler"
	configservice "thirdparty_service/internal/modules/config/service"
	cacheclient "thirdparty_service/internal/modules/execute/clients/cache_client"
	"thirdparty_service/internal/modules/execute/services"

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

	// Get the global tracer provider
	tracerProvider := otel.GetTracerProvider()
	if tracerProvider == nil {
		// If no tracer provider is set, use no-op provider
		tracerProvider = noop.NewTracerProvider()
		otel.SetTracerProvider(tracerProvider)
	}

	tracer := tracerProvider.Tracer(config.AppConfig.ServiceName)
	logger := otelslog.NewLogger(config.AppConfig.ServiceName)
	metricMeter := otel.Meter(config.AppConfig.ServiceName)
	cacheClient, err := cacheclient.NewRedisClient(config.AppConfig.CachingServiceGrpcURL)
	if err != nil {
		log.Println("Couldn't initalise the Cache client")
	}
	cacheManagerService, err := services.NewCacheManager(cacheClient)
	if err != nil {
		log.Println("Couldn't initalise the Cache client")
	}
	configService := configservice.NewConfigService(config.AppConfig.Environment, config.AppConfig.ServiceName, config.AppConfig.ConfigServiceURL, config.AppConfig.ConfigServiceUsername, config.AppConfig.ConfigServicePassword)
	configHandler := handler.NewConfigHandler(configService)
	
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

	if err := Container.Provide(func() configservice.ConfigService {
		return configService
	}); err != nil {
		return err
	}

	if err := Container.Provide(func() *handler.ConfigHandler {
		return configHandler
	}); err != nil {
		return err
	}
	if err := Container.Provide(func() *services.CacheManagerService {
		return cacheManagerService
	}); err != nil {
		return err
	}

	return nil
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
func GetConfigService() configservice.ConfigService {
	var configService configservice.ConfigService
	if err := Container.Invoke(func(c configservice.ConfigService) {
		configService = c
	}); err != nil {
		panic(err)
	}
	return configService
}

func GetConfigHandler() *handler.ConfigHandler {
	var configHandler *handler.ConfigHandler
	if err := Container.Invoke(func(c *handler.ConfigHandler) {
		configHandler = c
	}); err != nil {
		panic(err)
	}
	return configHandler
}

func GetCacheManagerService() *services.CacheManagerService {
	var cacheManagerService *services.CacheManagerService
	if err := Container.Invoke(func(c *services.CacheManagerService) {
		cacheManagerService = c
	}); err != nil {
		panic(err)
	}
	return cacheManagerService
}
