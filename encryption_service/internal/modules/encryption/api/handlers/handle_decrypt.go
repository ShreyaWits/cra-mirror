package handlers

import (
	"context"
	pb "encryption_microservice/internal/common/proto_gen"
	"encryption_microservice/internal/modules/encryption/api/dtos"
	"encryption_microservice/internal/modules/encryption/api/mapper"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleDecrypt handles the decryption request
func (h *EncryptionHandlerImpl) Decrypt(ctx context.Context, req *pb.DecryptRequest) (*pb.DecryptResponse, error) {

	token := req.Token

	userData, customErr := h.userService.GetUserData(token)
	if customErr != nil {
		return nil, status.Error(codes.Unauthenticated, customErr.Error())
	}

	mappedRequest := mapper.ConvertFromStructPB(req.Data)
	decryptRequest := &dtos.DecryptRequest{Data: mappedRequest}

	response, customErr := h.encryptionUseCase.Decrypt(ctx, userData.ID, userData.EDEKPrivate, userData.EDEKPublic, decryptRequest)
	if customErr != nil {
		return nil, status.Error(codes.Internal, customErr.Error())
	}

	mappedResponse := mapper.ConvertToStructPB(response.Data)
	return &pb.DecryptResponse{Data: mappedResponse}, nil
}
