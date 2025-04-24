package handlers

import (
	"context"
	pb "encryption_microservice/internal/common/proto_gen"
	"encryption_microservice/internal/modules/encryption/api/dtos"
	"encryption_microservice/internal/modules/encryption/api/mapper"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *EncryptionHandlerImpl) Encrypt(ctx context.Context, req *pb.EncryptDataRequest) (*pb.EncryptDataResponse, error) {
	token := req.Token

	userData, customErr := h.userService.GetUserData(token)
	if customErr != nil {
		return nil, status.Error(codes.Unauthenticated, customErr.Error())
	}

	mappedRequest := mapper.ConvertFromStructPB(req.Data)

	encryptedReq := &dtos.EncryptRequest{
		Data: mappedRequest,
	}

	// Execute use case
	response, customErr := h.encryptionUseCase.Encrypt(ctx, userData.ID, userData.EDEKPrivate, userData.EDEKPublic, encryptedReq)
	if customErr != nil {
		return nil, status.Error(codes.Internal, customErr.Error())
	}

	mappedResponse := mapper.ConvertToStructPB(response.Data)
	return &pb.EncryptDataResponse{
		Data: mappedResponse,
	}, nil
}
