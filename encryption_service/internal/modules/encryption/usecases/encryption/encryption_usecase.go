package usecases

import (
	"encryption_microservice/internal/modules/encryption/api/dtos"
	encryptionengine "encryption_microservice/internal/modules/encryption/services/encryption_engine"
	keymanager "encryption_microservice/internal/modules/encryption/services/key_manager"
	"encryption_microservice/internal/modules/encryption/services/user"
	"encryption_microservice/pkg/errors"
	"encryption_microservice/pkg/logger"
)

// EncryptionUseCaseImpl implements the EncryptionUseCase interface
type EncryptionUseCaseImpl struct {
	keyManager       keymanager.KeyManager
	encryptionEngine encryptionengine.EncryptionEngine
	userService      user.UserService
}

// NewEncryptionUseCase creates a new instance of EncryptionUseCaseImpl
func NewEncryptionUseCase(
	keyManager keymanager.KeyManager,
	encryptionEngine encryptionengine.EncryptionEngine,
	userService user.UserService,
) EncryptionUseCase {
	return &EncryptionUseCaseImpl{
		keyManager:       keyManager,
		encryptionEngine: encryptionEngine,
		userService:      userService,
	}
}

// Encrypt implements the encryption use case
func (u *EncryptionUseCaseImpl) Encrypt(req *dtos.EncryptRequest) (*dtos.EncryptResponse, error) {
	// Encrypt data and get EDEK
	encryptedData, err := u.encryptionEngine.Encrypt(req.Data, []byte(""))
	if err != nil {
		logger.Error("Failed to encrypt data", err)
		return nil, errors.NewEncryptionError("failed to encrypt data", err)
	}

	// Prepare response
	response := &dtos.EncryptResponse{
		EncryptedData: encryptedData,
	}

	return response, nil
}

// Decrypt implements the decryption use case
func (u *EncryptionUseCaseImpl) Decrypt(req *dtos.DecryptRequest) (*dtos.DecryptResponse, error) {
	// Decrypt data using EDEK
	decryptedData, err := u.encryptionEngine.Decrypt(req.EncryptedData, []byte(""))
	if err != nil {
		logger.Error("Failed to decrypt data", err)
		return nil, errors.NewEncryptionError("failed to decrypt data", err)
	}

	// Prepare response
	response := &dtos.DecryptResponse{
		Data: decryptedData,
	}

	return response, nil
}

// GenerateEDEK implements the EDEK generation use case
func (u *EncryptionUseCaseImpl) GenerateEDEK() (*dtos.GenerateEDEKResponse, error) {
	// Generate a new DEK
	dekPrivate, err := u.encryptionEngine.GenerateEncryptionKey()
	if err != nil {
		logger.Error("Failed to generate DEK Private", err)
		return nil, errors.NewEncryptionError("failed to generate DEK Private", err)
	}

	// Get KEK from key manager
	kekPrivate, err := u.keyManager.GenerateKEK()
	if err != nil {
		logger.Error("Failed to retrieve KEK Private", err)
		return nil, errors.NewKeyManagementError("KEKPrivate: failed to retrieve key encryption key", err)
	}

	user, err := u.userService.GetUserData("1")
	if err != nil {
		logger.Error("Failed to get user data", err)
		return nil, errors.NewUserServiceError("failed to get user data", err)
	}

	// Save KEKPrivate to KMS
	err = u.keyManager.StoreKEK(user.ID, kekPrivate)
	if err != nil {
		logger.Error("Failed to store KEKPrivate", err)
		return nil, errors.NewKeyManagementError("KEKPrivate: failed to store key encryption key", err)
	}

	// Generate EDEKPrivate
	edekPrivate, err := u.encryptionEngine.EncryptDEK(dekPrivate, kekPrivate)
	if err != nil {
		logger.Error("Failed to generate EDEKPrivate", err)
		return nil, errors.NewEncryptionError("EDEKPrivate: failed to generate encrypted key", err)
	}

	// Generate a new DEKPublic
	dekPublic, err := u.encryptionEngine.GenerateEncryptionKey()
	if err != nil {
		logger.Error("Failed to generate DEKPublic", err)
		return nil, errors.NewEncryptionError("failed to generate DEKPublic", err)
	}

	// Get KEK from key manager
	kekPublic, err := u.keyManager.GenerateKEK()
	if err != nil {
		logger.Error("Failed to retrieve KEKPublic", err)
		return nil, errors.NewKeyManagementError("KEKPublic: failed to retrieve key encryption key", err)
	}

	// Save KEKPublic to KMS
	if err := u.keyManager.StoreKEK("kekPublic", kekPublic); err != nil {
		logger.Error("Failed to store KEKPublic", err)
		return nil, errors.NewKeyManagementError("KEKPublic: failed to store key encryption key", err)
	}

	// Generate EDEKPublic
	edekPublic, err := u.encryptionEngine.EncryptDEK(dekPublic, kekPublic)
	if err != nil {
		logger.Error("Failed to generate EDEKPublic", err)
		return nil, errors.NewEncryptionError("EDEKPublic: failed to generate encrypted key", err)
	}

	// Prepare response
	response := &dtos.GenerateEDEKResponse{
		EDEKPrivate: edekPrivate,
		EDEKPublic:  edekPublic,
	}

	return response, nil
}
