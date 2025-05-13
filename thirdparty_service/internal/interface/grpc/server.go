package rpc

import (
	"thirdparty_service/internal/interface/grpc/methods"
	"thirdparty_service/internal/modules/execute/repositories"
	"thirdparty_service/internal/modules/execute/services"
	protos "thirdparty_service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func InitializeGRPCServer() (*grpc.Server, error) {
	grpcServer := grpc.NewServer()

	repository := repositories.New()
	services := services.New(repository)
	// register our service methods
	protos.RegisterThirdPartyServiceServer(grpcServer, methods.InitMethods(services))

	// enable server reflection
	reflection.Register(grpcServer)

	return grpcServer, nil
}
