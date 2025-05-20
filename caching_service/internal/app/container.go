package app

import (
	"log/slog"
	"redis-service/internal/repository"
	"redis-service/internal/service"
	"redis-service/pkg/config"
	"redis-service/pkg/redis"
	"strconv"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type Container struct {
	Tracer       trace.Tracer
	Logger       *slog.Logger
	Meter        metric.Meter
	RedisRepo    repository.RedisRepositoryInterface
	RedisService service.RedisServiceInterface
}

var Di *Container

func InitContainer() {
	// Get the global tracer provider
	tracerProvider := otel.GetTracerProvider()
	if tracerProvider == nil {
		// If no tracer provider is set, use no-op provider
		tracerProvider = noop.NewTracerProvider()
		otel.SetTracerProvider(tracerProvider)
	}

	tracer := tracerProvider.Tracer(config.SERVICE_NAME)
	logger := otelslog.NewLogger(config.SERVICE_NAME)
	metric := otel.Meter(config.SERVICE_NAME)

	redisDb, err := strconv.Atoi(config.REDIS_DB)
	if err != nil {
		logger.Error("Failed to parse REDIS_DB", "error", err)
		redisDb = 0 // Default to DB 0 on error
	}

	redisClient := redis.NewRedisService(config.REDIS_URL, config.REDIS_PASSWORD, redisDb)
	redisRepo := repository.NewRedisRepository(redisClient, logger)
	redisService := service.NewRedisService(redisRepo, logger)

	Di = &Container{
		Tracer:       tracer,
		Logger:       logger,
		Meter:        metric,
		RedisRepo:    redisRepo,
		RedisService: redisService,
	}
}
