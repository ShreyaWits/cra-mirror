package usecases

import (
	"encryption_microservice/internal/modules/encryption/api/dtos"
	"encryption_microservice/pkg/errors"
)

// EncryptionUseCase defines the interface for encryption operations
type EncryptionUseCase interface {
	// Encrypt encrypts data and returns encrypted data with EDEK
	Encrypt(userID string, edekPrivate string, edekPublic string, req *dtos.EncryptRequest) (*dtos.EncryptResponse, *errors.CustomError)

	// Decrypt decrypts data using EDEK and returns original data
	Decrypt(userID string, edekPrivate string, edekPublic string, req *dtos.DecryptRequest) (*dtos.DecryptResponse, *errors.CustomError)

	// GenerateEDEK generates an Encrypted DEK
	GenerateEDEK(userId string) (*dtos.GenerateEDEKResponse, *errors.CustomError)
}
