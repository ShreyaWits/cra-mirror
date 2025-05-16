package handlers

import (
	"context"
	pb "encryption_microservice/internal/common/proto_gen"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleGenerateEDEK handles the EDEK generation request
func (h *EncryptionHandlerImpl) GenerateEDEK(ctx context.Context, req *pb.GenerateEDEKRequest) (*pb.GenerateEDEKResponse, error) {
	token := req.Token

	userData, customErr := h.userService.GetUserData(token)
	if customErr != nil {
		return nil, status.Error(codes.Unauthenticated, customErr.Error())
	}

	// Execute use case
	response, customErr := h.encryptionUseCase.GenerateEDEK(ctx, userData.ID)
	if customErr != nil {
		return nil, status.Error(codes.Internal, customErr.Error())
	}

	fmt.Println(response, userData)

	// Fixme: Remove in Production as this will be executed by UserService
	if updateErr := h.userService.UpdateUser(userData.ID, response.EDEKPrivate, response.EDEKPublic); updateErr != nil {
		return nil, status.Error(codes.Internal, updateErr.Error())
	}

	return &pb.GenerateEDEKResponse{
		EDEKPrivate: response.EDEKPrivate,
		EDEKPublic:  response.EDEKPublic,
	}, nil
}
