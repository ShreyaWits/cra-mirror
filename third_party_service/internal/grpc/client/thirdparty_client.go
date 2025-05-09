package client

import (
	"log"
	pb "third_party_service/thirdparty-client-service/proto"

	"google.golang.org/grpc"
)

type ThirdPartyClient struct {
	Conn   *grpc.ClientConn
	Client pb.ThirdPartyServiceClient
}

func NewThirdPartyClient(address string) *ThirdPartyClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to gRPC server: %v", err)
	}

	return &ThirdPartyClient{
		Conn:   conn,
		Client: pb.NewThirdPartyServiceClient(conn),
	}
}
