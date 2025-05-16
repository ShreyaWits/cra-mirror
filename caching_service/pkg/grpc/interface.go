package grpc

import (
	"net"

	"google.golang.org/grpc"
)

type GRPCServerInterface interface {
	Serve(net.Listener) error
	GracefulStop()
	RegisterService(*grpc.ServiceDesc, interface{})
}

type GRPCListenerInterface interface {
	net.Listener
}
