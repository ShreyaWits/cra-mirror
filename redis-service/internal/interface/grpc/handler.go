package grpc

import (
	"context"
	"redis-service/internal/dto"
	"redis-service/internal/repository"
	"redis-service/internal/service"
	"redis-service/internal/utils"
	"redis-service/pkg/message"
	"redis-service/pkg/redis"
	"redis-service/proto"
)

type GRPCServer struct {
	proto.UnimplementedCacheServiceServer
	service service.RedisServiceInterface
}

func NewGRPCHandler() *GRPCServer {
	// Initialize the Redis repository and service
	redisClient := redis.NewRedisService("localhost:6379", "password", 0)
	redisRepo := repository.NewRedisRepository(redisClient)
	redisService := service.NewRedisService(redisRepo)
	return &GRPCServer{service: redisService}
}

func (grpc *GRPCServer) GetCache(ctx context.Context, req *proto.GetCacheRequest) (*proto.GetCacheResponse, error) {

	payload := dto.GetCacheRequest{
		Namespace:  req.Namespace,
		Key:        req.Key,
		TrackingId: req.TrackingId,
	}
	err := utils.ValidateStruct(payload)
	if err != nil {
		return &proto.GetCacheResponse{
			Error:   err.Error(),
			Found:   false,
			Message: message.RD0001,
		}, nil
	}

	value, err := grpc.service.GetCache(&payload)
	if err != nil || value == "" {
		return &proto.GetCacheResponse{Value: value, Found: false, Message: message.RD0002}, nil
	}
	return &proto.GetCacheResponse{Value: value, Found: true, Message: message.RD0000}, nil

}
func (grpc *GRPCServer) SetCache(ctx context.Context, req *proto.SetCacheRequest) (*proto.SetCacheResponse, error) {
	payload := dto.SetCacheRequest{
		Namespace:  req.Namespace,
		Key:        req.Key,
		TrackingId: req.TrackingId,
		Value:      req.Value,
		TTL:        req.Ttl,
	}
	err := utils.ValidateStruct(payload)
	if err != nil {
		return &proto.SetCacheResponse{
			Success: false,
			Message: message.RD0004,
		}, nil
	}
	value, err := grpc.service.SetCache(&payload)
	if err != nil || !value {
		return &proto.SetCacheResponse{Success: false, Message: message.RD0002}, nil
	}
	return &proto.SetCacheResponse{Success: true, Message: message.RD0000}, nil

}
func (grpc *GRPCServer) InvalidateCache(ctx context.Context, req *proto.InvalidateCacheRequest) (*proto.InvalidateCacheResponse, error) {
	payload := dto.DeleteCacheRequest{
		Namespace:  req.Namespace,
		Key:        req.Key,
		TrackingId: req.TrackingId,
	}
	err := utils.ValidateStruct(payload)
	if err != nil {
		return &proto.InvalidateCacheResponse{
			Success: false,
			Message: message.RD0004,
		}, nil
	}
	value, err := grpc.service.InvalidateCache(&payload)
	if err != nil || !value {
		return &proto.InvalidateCacheResponse{Success: false, Message: message.RD0002}, nil
	}
	return &proto.InvalidateCacheResponse{Success: true, Message: message.RD0000}, nil

}
