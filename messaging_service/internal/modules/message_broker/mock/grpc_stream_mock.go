// MockSubscribeServer is a mock implementation of pb.MessagingService_SubscribeV1Server

package mock

import (
	context "context"

	pb "cra-protos/messaging_service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// NewMockSubscribeServer creates a new instance of MockSubscribeServer with the given context
func NewMockSubscribeServer(ctx context.Context) *MockSubscribeServer {
	return &MockSubscribeServer{
		ctx: ctx,
	}
}

type MockSubscribeServer struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *MockSubscribeServer) Context() context.Context {
	return m.ctx
}

func (m *MockSubscribeServer) Send(msg *pb.KafkaMessage) error {
	return nil
}

func (m *MockSubscribeServer) RecvMsg(interface{}) error {
	return nil
}

func (m *MockSubscribeServer) SendHeader(md metadata.MD) error {
	return nil
}

func (m *MockSubscribeServer) SetHeader(md metadata.MD) error {
	return nil
}

func (m *MockSubscribeServer) SetTrailer(md metadata.MD) {
}
