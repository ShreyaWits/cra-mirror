package usecases

import (
	"context"
	"encryption_microservice/internal/modules/encryption/api/dtos"
	"encryption_microservice/pkg/errors"
)

// EncryptionUseCase defines the interface for encryption operations
type EncryptionUseCase interface {
	// Encrypt encrypts data and returns encrypted data with EDEK
	Encrypt(context context.Context, userID string, edekPrivate string, edekPublic string, req *dtos.EncryptRequest) (*dtos.EncryptResponse, *errors.CustomError)

	// Decrypt decrypts data using EDEK and returns original data
	Decrypt(context context.Context, userID string, edekPrivate string, edekPublic string, req *dtos.DecryptRequest) (*dtos.DecryptResponse, *errors.CustomError)

	// GenerateEDEK generates an Encrypted DEK
	GenerateEDEK(context context.Context, userId string) (*dtos.GenerateEDEKResponse, *errors.CustomError)
}
