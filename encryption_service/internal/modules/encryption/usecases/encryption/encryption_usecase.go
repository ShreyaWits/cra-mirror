package usecases

import (
	"encryption_microservice/internal/modules/encryption/api/dtos"
	enums "encryption_microservice/internal/modules/encryption/api/enum"
	encryptionengine "encryption_microservice/internal/modules/encryption/services/encryption_engine"
	keymanager "encryption_microservice/internal/modules/encryption/services/key_manager"
	"encryption_microservice/pkg/errors"
	"encryption_microservice/pkg/logger"
	"fmt"
)

// EncryptionUseCaseImpl implements the EncryptionUseCase interface
type EncryptionUseCaseImpl struct {
	keyManager       keymanager.KeyManager
	encryptionEngine encryptionengine.EncryptionEngine
}

// NewEncryptionUseCase creates a new instance of EncryptionUseCaseImpl
func NewEncryptionUseCase(
	keyManager keymanager.KeyManager,
	encryptionEngine encryptionengine.EncryptionEngine) EncryptionUseCase {
	return &EncryptionUseCaseImpl{
		keyManager:       keyManager,
		encryptionEngine: encryptionEngine,
	}
}

func (u *EncryptionUseCaseImpl) getDek(kekID string, edek string) ([]byte, error) {
	kek, err := u.keyManager.RetrieveKEK(kekID)
	if err != nil {
		logger.Error("Failed to retrieve private KEK", err)
		return nil, errors.NewEncryptionError("failed to retrieve private KEK", err)
	}
	dek, err := u.encryptionEngine.DecryptDEK(edek, kek)
	if err != nil {
		logger.Error("Failed to retrieve private DEK", err)
		return nil, errors.NewEncryptionError("failed to retrieve private DEK", err)
	}
	return dek, nil
}

// Encrypt implements the encryption use case
func (u *EncryptionUseCaseImpl) Encrypt(userID string, edekPrivate, edekPublic string, req *dtos.EncryptRequest) (*dtos.EncryptResponse, error) {
	var (
		DEKMap         = make(map[enums.KeyType][]byte)
		retrievedFlags = make(map[enums.KeyType]bool)
		encryptedItems []map[string]interface{}
	)

	for _, item := range req.Data {
		eType, ok := item["e_type"].(string)
		isValid, keyType := enums.IsValidKeyType(eType)
		if !ok || !isValid {
			err := fmt.Errorf("e_type is missing or invalid")
			logger.Error("Invalid encryption type in item", err)
			return nil, err
		}

		// Fetch DEK if not already fetched
		if !retrievedFlags[keyType] {
			kekID := keyType.AddKeyTypeIdentifier(userID)
			var err error
			switch keyType {
			case enums.KeyPrivate:
				DEKMap[keyType], err = u.getDek(kekID, edekPrivate)
			case enums.KeyPublic:
				DEKMap[keyType], err = u.getDek(kekID, edekPublic)
			default:
				err = fmt.Errorf("unsupported KMS type: %s", eType)
			}
			if err != nil {
				logger.Error("Failed to retrieve DEK", err)
				return nil, errors.NewEncryptionError("failed to retrieve DEK", err)
			}
			retrievedFlags[keyType] = true
		}

		// Encrypt the item using the proper DEK
		encryptItem, err := u.encryptItemFields(item, DEKMap[keyType], keyType)
		if err != nil {
			logger.Error("Failed to encrypt item", err)
			return nil, errors.NewEncryptionError("failed to encrypt item", err)
		}
		encryptedItems = append(encryptedItems, encryptItem)
	}

	return &dtos.EncryptResponse{Data: encryptedItems}, nil
}

func (u *EncryptionUseCaseImpl) encryptItemFields(item map[string]interface{}, dek []byte, prefix enums.KeyType) (map[string]interface{}, error) {
	encryptedItem := make(map[string]interface{})

	for key, value := range item {
		if key == "e_type" {
			continue // Skip e_type
		}
		strVal := fmt.Sprintf("%v", value)
		encryptedVal, err := u.encryptionEngine.Encrypt(strVal, dek)
		if err != nil {
			logger.Error("Failed to encrypt value", err)
			return nil, errors.NewEncryptionError(fmt.Sprintf("failed to encrypt field: %s", key), err)
		}
		encryptedItem[key] = prefix.AddKeyTypeIdentifier(encryptedVal)
	}

	return encryptedItem, nil
}

func (u *EncryptionUseCaseImpl) Decrypt(userID string, edekPrivate string, edekPublic string, req *dtos.DecryptRequest) (*dtos.DecryptResponse, error) {
	var (
		DEKMap         = make(map[enums.KeyType][]byte)
		retrievedFlags = make(map[enums.KeyType]bool)
		decryptedItems = make([]map[string]interface{}, 0, len(req.Data))
	)

	for _, item := range req.Data {
		decryptedItem := make(map[string]interface{})

		for key, value := range item {
			strVal := fmt.Sprintf("%v", value)

			// Skip if no KMS identifier
			if !enums.HasKeyTypeIdentifier(strVal) {
				logger.Error("Skipping field without recognized prefix", fmt.Errorf("field: %s", key))
				decryptedItem[key] = strVal
				continue
			}

			encryptedString, keyTypePtr, _ := enums.RemoveKeyTypeIdentifier(strVal)
			keyType := *keyTypePtr

			var dek []byte
			var err error

			// Retrieve DEK if not already cached
			if !retrievedFlags[keyType] {
				kekID := keyType.AddKeyTypeIdentifier(userID)
				edek := edekPrivate
				if keyType == enums.KeyPublic {
					edek = edekPublic
				}

				dek, err = u.getDek(kekID, edek)
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to retrieve %v DEK", keyType), err)
					return nil, errors.NewEncryptionError(fmt.Sprintf("failed to retrieve %v DEK", keyType), err)
				}
				DEKMap[keyType] = dek
				retrievedFlags[keyType] = true
			} else {
				dek = DEKMap[keyType]
			}

			// Decrypt field
			plainText, err := u.encryptionEngine.Decrypt(encryptedString, dek)
			if err != nil {
				logger.Error("Failed to decrypt field", err)
				return nil, errors.NewEncryptionError(fmt.Sprintf("failed to decrypt field: %s", key), err)
			}
			decryptedItem[key] = plainText
		}

		decryptedItems = append(decryptedItems, decryptedItem)
	}

	return &dtos.DecryptResponse{
		Data: decryptedItems,
	}, nil
}

// GenerateEDEK implements the EDEK generation use case
func (u *EncryptionUseCaseImpl) GenerateEDEK(userID string) (*dtos.GenerateEDEKResponse, error) {
	// Helper function to generate, store KEK and return EDEK
	generateEDEK := func(keyType enums.KeyType) (string, error) {
		// Generate DEK
		dek, err := u.encryptionEngine.GenerateEncryptionKey()
		if err != nil {
			logger.Error("Failed to generate DEK for", err)
			return "", errors.NewEncryptionError("failed to generate DEK", err)
		}

		// Generate KEK
		kek, err := u.encryptionEngine.GenerateEncryptionKey()
		if err != nil {
			logger.Error("Failed to generate KEK for", err)
			return "", errors.NewKeyManagementError("failed to generate KEK", err)
		}

		// Store KEK in KMS
		kmsKeyName := keyType.AddKeyTypeIdentifier(userID)
		if err := u.keyManager.StoreKEK(kmsKeyName, kek); err != nil {
			logger.Error("Failed to store KEK for", err)
			return "", errors.NewKeyManagementError("failed to store KEK", err)
		}

		// Encrypt DEK to generate EDEK
		edek, err := u.encryptionEngine.EncryptDEK(dek, kek)
		if err != nil {
			logger.Error("Failed to encrypt DEK (EDEK) for", err)
			return "", errors.NewEncryptionError("failed to encrypt DEK", err)
		}

		return edek, nil
	}

	// Generate EDEKs for both private and public
	edekPrivate, err := generateEDEK(enums.KeyPrivate)
	if err != nil {
		return nil, err
	}

	edekPublic, err := generateEDEK(enums.KeyPublic)
	if err != nil {
		return nil, err
	}

	return &dtos.GenerateEDEKResponse{
		EDEKPrivate: edekPrivate,
		EDEKPublic:  edekPublic,
	}, nil
}
