package rpc

import (
	"context"
	"thirdparty_service/internal/interface/grpc/methods"
	"thirdparty_service/internal/modules/execute/repositories"
	"thirdparty_service/internal/modules/execute/services"
	protos "thirdparty_service/proto"

	push_service "thirdparty_service/pkg/push"
	"thirdparty_service/pkg/sendgrid"
	twilio_sms "thirdparty_service/pkg/twilio"
	"thirdparty_service/pkg/whatsapp"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func InitializeGRPCServer() (*grpc.Server, error) {
	grpcServer := grpc.NewServer()

	repository := repositories.New()
	services := services.New(
		repository,
		twilio_sms.NewTwilioClient,
		sendgrid.NewSendGridClient,
		whatsapp.NewWhatsAppClient,
		func(ctx context.Context, creds, projectID string) (push_service.FCMClient, error) {
			client, err := push_service.NewFCMClient(ctx, creds, projectID)
			if err != nil {
				return push_service.FCMClient{}, err
			}
			return *client, nil
		},
	)
	// register our service methods
	protos.RegisterThirdPartyServiceServer(grpcServer, methods.InitMethods(services))

	// enable server reflection
	reflection.Register(grpcServer)

	return grpcServer, nil
}
