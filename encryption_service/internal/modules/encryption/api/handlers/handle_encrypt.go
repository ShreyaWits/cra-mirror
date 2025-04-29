package handlers

import (
	"context"
	pb "encryption_microservice/internal/common/proto_gen"
	"encryption_microservice/internal/modules/encryption/api/dtos"
	"encryption_microservice/internal/modules/encryption/api/mapper"
	"encryption_microservice/pkg/logger"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *EncryptionHandlerImpl) Encrypt(ctx context.Context, req *pb.EncryptDataRequest) (*pb.EncryptDataResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Failed to retrieve private KEK", fmt.Errorf("%v", r))
			status.Error(codes.Internal, fmt.Sprintf("%v", r))
		}
	}()

	token := req.Token

	userData, customErr := h.userService.GetUserData(token)
	if customErr != nil {
		return nil, status.Error(codes.Unauthenticated, customErr.Error())
	}
	var keyId string
	if req.UserId == nil || *req.UserId == "" {
		keyId = userData.ID
	} else {
		userData, customErr = h.userService.GetUserData(*req.UserId)
		if customErr != nil {
			return nil, status.Error(codes.Unauthenticated, customErr.Error())
		}
		keyId = userData.ID
	}
	mappedRequest := mapper.ConvertFromStructPB(req.Data)

	encryptedReq := &dtos.EncryptRequest{
		Data: mappedRequest,
	}
	fmt.Print("*******Key", keyId)

	// Execute use case
	response, customErr := h.encryptionUseCase.Encrypt(ctx, keyId, userData.EDEKPrivate, userData.EDEKPublic, encryptedReq)
	if customErr != nil {
		return nil, status.Error(codes.Internal, customErr.Error())
	}

	mappedResponse := mapper.ConvertToStructPB(response.Data)
	return &pb.EncryptDataResponse{
		Data: mappedResponse,
	}, nil
}
