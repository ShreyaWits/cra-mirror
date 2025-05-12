package rpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"thirdparty_service/api/rpc/methods"
	"thirdparty_service/internal/pb"
)

func InitializeGRPCServer() (*grpc.Server, error) {
	grpcServer := grpc.NewServer()

	// register our service methods
	pb.RegisterHelloWorldServer(grpcServer, &methods.HelloWorldServer{})

	// enable server reflection
	reflection.Register(grpcServer)

	return grpcServer, nil
}
