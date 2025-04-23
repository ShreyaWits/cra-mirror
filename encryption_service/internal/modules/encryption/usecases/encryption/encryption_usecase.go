package usecases

import (
	"context"
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

func (u *EncryptionUseCaseImpl) getDek(kekID string, edek string) ([]byte, *errors.CustomError) {
	kek, err := u.keyManager.RetrieveKEK(kekID)
	if err != nil {
		logger.Error("Failed to retrieve private KEK", err.ErrObj())
		return nil, err
	}
	fmt.Println("Keys to Decrypt ", edek, kek)
	dek, err := u.encryptionEngine.DecryptDEK(edek, kek)
	if err != nil {
		logger.Error("Failed to retrieve DEK", err.ErrObj())
		return nil, err
	}
	return dek, nil
}

// Encrypt implements the encryption use case
func (u *EncryptionUseCaseImpl) Encrypt(context context.Context, userID string, edekPrivate, edekPublic string, req *dtos.EncryptRequest) (*dtos.EncryptResponse, *errors.CustomError) {
	var (
		DEKMap         = make(map[enums.KeyType][]byte)
		retrievedFlags = make(map[enums.KeyType]bool)
		encryptedItems []map[string]string
	)

	for _, item := range req.Data {
		eType, ok := item["e_type"]
		isValid, keyType := enums.IsValidKeyType(eType)
		if !ok || !isValid {
			err := fmt.Errorf("e_type is missing or invalid")
			return nil, errors.NewCustomError(errors.ESErrETYPEKeyMissing, err)
		}

		// Fetch DEK if not already fetched
		if !retrievedFlags[keyType] {
			kekID := keyType.AddKeyTypeIdentifier(userID)
			var err *errors.CustomError
			switch keyType {
			case enums.KeyPrivate:
				DEKMap[keyType], err = u.getDek(kekID, edekPrivate)
			case enums.KeyPublic:
				DEKMap[keyType], err = u.getDek(kekID, edekPublic)
			default:
				err = errors.NewCustomError(errors.ESErrRetrieveKEK, fmt.Errorf("unsupported KMS type: %s", eType))
			}
			if err != nil {
				logger.Error("Failed to retrieve DEK", err.ErrObj())
				return nil, err
			}
			retrievedFlags[keyType] = true
		}

		// Encrypt the item using the proper DEK
		encryptItem, err := u.encryptItemFields(item, DEKMap[keyType], keyType)
		if err != nil {
			logger.Error("Failed to encrypt item", err.ErrObj())
			return nil, err
		}
		encryptedItems = append(encryptedItems, encryptItem)
	}

	return &dtos.EncryptResponse{Data: encryptedItems}, nil
}

func (u *EncryptionUseCaseImpl) encryptItemFields(item map[string]string, dek []byte, prefix enums.KeyType) (map[string]string, *errors.CustomError) {
	encryptedItem := make(map[string]string)

	for key, value := range item {
		if key == "e_type" {
			continue // Skip e_type
		}
		strVal := fmt.Sprintf("%v", value)
		encryptedVal, err := u.encryptionEngine.Encrypt(strVal, dek)
		if err != nil {
			logger.Error("Failed to encrypt value", err.ErrObj())
			return nil, err
		}
		encryptedItem[key] = prefix.AddKeyTypeIdentifier(encryptedVal)
	}

	return encryptedItem, nil
}

func (u *EncryptionUseCaseImpl) Decrypt(context context.Context, userID string, edekPrivate string, edekPublic string, req *dtos.DecryptRequest) (*dtos.DecryptResponse, *errors.CustomError) {
	var (
		DEKMap         = make(map[enums.KeyType][]byte)
		retrievedFlags = make(map[enums.KeyType]bool)
		decryptedItems = make([]map[string]string, 0, len(req.Data))
	)

	for _, item := range req.Data {
		decryptedItem := make(map[string]string)

		for key, value := range item {
			strVal := fmt.Sprintf("%v", value)

			// Skip if no KMS identifier
			if !enums.HasKeyTypeIdentifier(strVal) {
				decryptedItem[key] = strVal
				continue
			}

			encryptedString, keyTypePtr, _ := enums.RemoveKeyTypeIdentifier(strVal)
			keyType := *keyTypePtr

			var dek []byte
			var err *errors.CustomError

			// Retrieve DEK if not already cached
			if !retrievedFlags[keyType] {
				kekID := keyType.AddKeyTypeIdentifier(userID)
				edek := edekPrivate
				if keyType == enums.KeyPublic {
					edek = edekPublic
				}

				dek, err = u.getDek(kekID, edek)
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to retrieve %v DEK", keyType), err.ErrObj())
					return nil, err
				}
				DEKMap[keyType] = dek
				retrievedFlags[keyType] = true
			} else {
				dek = DEKMap[keyType]
			}

			// Decrypt field
			plainText, err := u.encryptionEngine.Decrypt(encryptedString, dek)
			if err != nil {
				logger.Error("Failed to decrypt field", err.ErrObj())
				return nil, err
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
func (u *EncryptionUseCaseImpl) GenerateEDEK(context context.Context, userID string) (*dtos.GenerateEDEKResponse, *errors.CustomError) {
	// Helper function to generate, store KEK and return EDEK
	generateEDEK := func(keyType enums.KeyType) (string, *errors.CustomError) {
		// Generate DEK
		dek, err := u.encryptionEngine.GenerateEncryptionKey()
		if err != nil {
			logger.Error("Failed to generate DEK for", err.ErrObj())
			return "", err
		}

		// Generate KEK
		kek, err := u.encryptionEngine.GenerateEncryptionKey()
		if err != nil {
			logger.Error("Failed to generate KEK for", err.ErrObj())
			return "", err
		}

		// Store KEK in KMS
		kmsKeyName := keyType.AddKeyTypeIdentifier(userID)
		if err := u.keyManager.StoreKEK(kmsKeyName, kek); err != nil {
			logger.Error("Failed to store KEK for", err.ErrObj())
			return "", err
		}

		// Encrypt DEK to generate EDEK
		edek, err := u.encryptionEngine.EncryptDEK(dek, kek)
		if err != nil {
			logger.Error("Failed to encrypt DEK (EDEK) for", err.ErrObj())
			return "", err
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
