package rpc

import (
	"thirdparty_service/api/rpc/methods"
	protos "thirdparty_service/protos"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func InitializeGRPCServer() (*grpc.Server, error) {
	grpcServer := grpc.NewServer()

	// register our service methods
	// pb.RegisterHelloWorldServer(grpcServer, &methods.HelloWorldServer{})
	protos.RegisterThirdPartyServiceServer(grpcServer, &methods.ThirdPartyServer{})

	// enable server reflection
	reflection.Register(grpcServer)

	return grpcServer, nil
}
