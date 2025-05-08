package grpc

import (
	"context"
	"redis-service/internal/repository"
	"redis-service/internal/service"
	"redis-service/proto"
)

type GRPCServer struct {
	proto.UnimplementedCacheServiceServer
	service service.RedisServiceInterface
}

func NewGRPCHandler() *GRPCServer {
	// Initialize the Redis repository and service
	redisRepo := repository.NewRedisRepository()
	redisService := service.NewRedisService(redisRepo)
	return &GRPCServer{service: redisService}
}

func (grpc *GRPCServer) GetCache(ctx context.Context, req *proto.GetCacheRequest) (*proto.GetCacheResponse, error) {

	key := req.Key
	value, err := grpc.service.GetCache(key)
	if err != nil {
		return nil, err
	}
	return &proto.GetCacheResponse{Value: value}, nil

}
