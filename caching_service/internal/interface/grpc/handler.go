package grpc

import (
	"context"
	"log"
	"log/slog"
	"redis-service/internal/app"
	"redis-service/internal/dto"
	"redis-service/internal/repository"
	"redis-service/internal/service"
	"redis-service/internal/utils"
	"redis-service/pkg/config"
	"redis-service/pkg/message"
	"redis-service/pkg/redis"
	"redis-service/proto"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	proto.UnimplementedCacheServiceServer
	service service.RedisServiceInterface
	tracer  trace.Tracer
	logger  *slog.Logger
}

func NewGRPCHandler() *GRPCServer {
	// Initialize the Redis repository and service

	redisDb, err := strconv.Atoi(config.REDIS_DB)
	if err != nil {
		redisDb = 0
	}

	tracer := app.Di.Tracer
	logger := app.Di.Logger

	redisClient := redis.NewRedisService(config.REDIS_URL, config.REDIS_PASSWORD, redisDb)
	redisRepo := repository.NewRedisRepository(redisClient, logger)
	redisService := service.NewRedisService(redisRepo, logger)
	return &GRPCServer{service: redisService, tracer: tracer, logger: logger}
}

func (grpc *GRPCServer) GetCache(ctx context.Context, req *proto.GetCacheRequest) (*proto.GetCacheResponse, error) {

	// Get span from context if available
	var span trace.Span
	if grpc.tracer != nil {
		span = trace.SpanFromContext(ctx)
	}

	grpc.logger.InfoContext(ctx, "get cache request", "namespace", req.Namespace, "key", req.Key)

	// Add span attributes if we have a span
	if span != nil {
		span.SetAttributes(
			attribute.String("cache.namespace", req.Namespace),
			attribute.String("cache.key", req.Key),
		)
	}

	// Convert request to DTO
	payload := &dto.GetCacheRequest{
		Namespace: req.Namespace,
		Key:       req.Key,
	}

	// Validate request
	if err := utils.ValidateStruct(payload); err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
		}
		grpc.logger.Error("set cache validation failed",
			"error", err.Error(),
			"namespace", req.Namespace,
			"key", req.Key,
		)
		return &proto.GetCacheResponse{
			Found:   false,
			Message: message.RD0004,
		}, nil
	}

	// Get cache value
	value, err := grpc.service.GetCache(payload)
	if err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
		}
		grpc.logger.Error("failed to get cache",
			"error", err.Error(),
			"namespace", req.Namespace,
			"key", req.Key,
		)
		return nil, status.Error(grpccodes.Internal, err.Error())
	}

	// Log success
	grpc.logger.Info("cache retrieved successfully",
		"namespace", req.Namespace,
		"key", req.Key,
	)

	return &proto.GetCacheResponse{
		Value: value,
		Found: value != "",
	}, nil
}

func (grpc *GRPCServer) SetCache(ctx context.Context, req *proto.SetCacheRequest) (*proto.SetCacheResponse, error) {

	// Get span from context if available
	var span trace.Span
	if grpc.tracer != nil {
		span = trace.SpanFromContext(ctx)
	}

	// Log request
	grpc.logger.InfoContext(ctx, "set cache request",
		"namespace", req.Namespace,
		"key", req.Key,
		"ttl", req.Ttl,
	)

	// Add span attributes if we have a span
	if span != nil {
		span.SetAttributes(
			attribute.String("cache.namespace", req.Namespace),
			attribute.String("cache.key", req.Key),
			attribute.Int64("cache.ttl", req.Ttl),
		)
	}

	// Convert request to DTO
	payload := &dto.SetCacheRequest{
		Namespace: req.Namespace,
		Key:       req.Key,
		Value:     req.Value,
		TTL:       req.Ttl,
	}

	// Validate request
	if err := utils.ValidateStruct(payload); err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
		}
		grpc.logger.Error("set cache validation failed",
			"error", err.Error(),
			"namespace", req.Namespace,
			"key", req.Key,
		)
		return &proto.SetCacheResponse{
			Success: false,
			Message: message.RD0004,
		}, nil
	}

	// Set cache value
	success, err := grpc.service.SetCache(payload)

	log.Println("err", err, success)

	if err != nil || !success {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, "failed to set cache")
		}
		grpc.logger.Error("set cache failed",
			"error", err.Error(),
			"namespace", req.Namespace,
			"key", req.Key,
		)
		return &proto.SetCacheResponse{
			Success: false,
			Message: message.RD0003,
		}, nil
	}

	// Log success
	grpc.logger.Info("set cache successful",
		"namespace", req.Namespace,
		"key", req.Key,
	)

	return &proto.SetCacheResponse{
		Success: true,
		Message: message.RD0000,
	}, nil
}

func (grpc *GRPCServer) InvalidateCache(ctx context.Context, req *proto.InvalidateCacheRequest) (*proto.InvalidateCacheResponse, error) {

	// Get span from context if available
	var span trace.Span
	if grpc.tracer != nil {
		span = trace.SpanFromContext(ctx)
	}

	// Log request
	grpc.logger.Info("invalidate cache request",
		"namespace", req.Namespace,
		"key", req.Key,
	)

	// Add span attributes if we have a span
	if span != nil {
		span.SetAttributes(
			attribute.String("cache.namespace", req.Namespace),
			attribute.String("cache.key", req.Key),
		)
	}

	// Convert request to DTO
	payload := &dto.DeleteCacheRequest{
		Namespace: req.Namespace,
		Key:       req.Key,
	}

	// Validate request
	if err := utils.ValidateStruct(payload); err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
		}
		grpc.logger.Error("invalidate cache validation failed",
			"error", err.Error(),
			"namespace", req.Namespace,
			"key", req.Key,
		)
		return &proto.InvalidateCacheResponse{
			Success: false,
			Message: message.RD0004,
		}, nil
	}

	// Invalidate cache
	success, err := grpc.service.InvalidateCache(payload)
	if err != nil || !success {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, "failed to invalidate cache")
		}
		grpc.logger.Error("invalidate cache failed",
			"error", err.Error(),
			"namespace", req.Namespace,
			"key", req.Key,
		)
		return &proto.InvalidateCacheResponse{
			Success: false,
			Message: message.RD0002,
		}, nil
	}

	// Log success
	grpc.logger.Info("invalidate cache successful",
		"namespace", req.Namespace,
		"key", req.Key,
	)

	return &proto.InvalidateCacheResponse{
		Success: true,
		Message: message.RD0000,
	}, nil
}
