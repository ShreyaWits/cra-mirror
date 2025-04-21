package handlers

import usecases "encryption_microservice/internal/modules/encryption/usecases/encryption"

// EncryptionHandlerImpl implements the EncryptionHandler interface
type EncryptionHandlerImpl struct {
	encryptionUseCase usecases.EncryptionUseCase
}

// NewEncryptionHandler creates a new instance of EncryptionHandlerImpl
func NewEncryptionHandler(
	encryptionUseCase usecases.EncryptionUseCase,
) *EncryptionHandlerImpl {
	return &EncryptionHandlerImpl{
		encryptionUseCase: encryptionUseCase,
	}
}
