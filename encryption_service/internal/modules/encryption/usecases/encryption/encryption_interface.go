package usecases

import (
	"encryption_microservice/internal/modules/encryption/api/dtos"
)

// EncryptionUseCase defines the interface for encryption operations
type EncryptionUseCase interface {
	// Encrypt encrypts data and returns encrypted data with EDEK
	Encrypt(req *dtos.EncryptRequest) (*dtos.EncryptResponse, error)

	// Decrypt decrypts data using EDEK and returns original data
	Decrypt(req *dtos.DecryptRequest) (*dtos.DecryptResponse, error)

	// GenerateEDEK generates an Encrypted DEK
	GenerateEDEK() (*dtos.GenerateEDEKResponse, error)
}
