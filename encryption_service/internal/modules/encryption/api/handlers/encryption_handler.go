package handlers

import (
	pb "encryption_microservice/internal/common/proto_gen"
	"encryption_microservice/internal/modules/encryption/services/user"
	usecases "encryption_microservice/internal/modules/encryption/usecases/encryption"
)

// EncryptionHandlerImpl implements the EncryptionHandler interface
type EncryptionHandlerImpl struct {
	pb.UnimplementedEncryptionServiceServer
	encryptionUseCase usecases.EncryptionUseCase
	userService       user.UserService
}

// NewEncryptionHandler creates a new instance of EncryptionHandlerImpl
func NewEncryptionHandler(
	encryptionUseCase usecases.EncryptionUseCase,
	userService user.UserService,
) *EncryptionHandlerImpl {
	return &EncryptionHandlerImpl{
		encryptionUseCase: encryptionUseCase,
		userService:       userService,
	}
}
