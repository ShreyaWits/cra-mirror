package grpc

import (
	"net"
)

type GRPCServerInterface interface {
	Serve(net.Listener) error
	GracefulStop()
}

type GRPCListenerInterface interface {
	net.Listener
}
