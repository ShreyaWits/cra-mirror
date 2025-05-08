package grpc

import (
	"context"
	"redis-service/proto"
)

type GRPCHandlerInterface interface {
	GetCache(context.Context, *proto.GetCacheRequest) (*proto.GetCacheResponse, error)
}
