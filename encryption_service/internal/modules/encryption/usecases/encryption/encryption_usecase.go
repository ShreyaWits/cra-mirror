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
	encryptionEngine encryptionengine.EncryptionEngine,
) EncryptionUseCase {
	return &EncryptionUseCaseImpl{
		keyManager:       keyManager,
		encryptionEngine: encryptionEngine,
	}
}

// Encrypt implements the encryption use case
func (u *EncryptionUseCaseImpl) Encrypt(userID string, edekPrivate string, edekPublic string, req *dtos.EncryptRequest) (*dtos.EncryptResponse, error) {

	var DEKPrivate, DEKPublic []byte
	var encryptedItems []map[string]interface{}
	privateRetrieved := false
	publicRetrieved := false

	// Validate and prepare KEKs
	for _, item := range req.Data {
		eType, ok := item["e_type"].(string)
		isValid, kmsType := enums.IsValidKMSType(eType)
		if !ok || !isValid {
			logger.Error("Missing or invalid 'e_type' in item", fmt.Errorf("e_type is not valid"))
			return nil, fmt.Errorf("missing or invalid 'e_type' in item")
		}

		kekID := kmsType.AddKMSTypeIdentifier(userID)

		if kmsType == enums.KMSPrivate {
			if !privateRetrieved {
				kek, err := u.keyManager.RetrieveKEK(kekID)
				if err != nil {
					logger.Error("Failed to retrieve private KEK", err)
					return nil, errors.NewEncryptionError("failed to retrieve private KEK", err)
				}
				DEKPrivate, err = u.encryptionEngine.DecryptDEK(edekPrivate, kek)
				if err != nil {
					logger.Error("Failed to retrieve private DEK", err)
					return nil, errors.NewEncryptionError("failed to retrieve private DEK", err)
				}
				privateRetrieved = true
			}
			encryptItem, err := u.encryptItemFields(item, DEKPrivate, kmsType)
			if err != nil {
				logger.Error("Failed to encrypt item", err)
				return nil, errors.NewEncryptionError("failed to encrypt item", err)
			}
			encryptedItems = append(encryptedItems, encryptItem)
		}

		if kmsType == enums.KMSPublic {
			if !publicRetrieved {
				kek, err := u.keyManager.RetrieveKEK(kekID)
				if err != nil {
					logger.Error("Failed to retrieve public KEK", err)
					return nil, errors.NewEncryptionError("failed to retrieve public KEK", err)
				}
				DEKPublic, err = u.encryptionEngine.DecryptDEK(edekPublic, kek)
				if err != nil {
					logger.Error("Failed to retrieve public DEK", err)
					return nil, errors.NewEncryptionError("failed to retrieve public DEK", err)
				}
				publicRetrieved = true

			}
			encryptItem, err := u.encryptItemFields(item, DEKPublic, kmsType)
			if err != nil {
				logger.Error("Failed to encrypt item", err)
				return nil, errors.NewEncryptionError("failed to encrypt item", err)
			}
			encryptedItems = append(encryptedItems, encryptItem)
		}
	}

	return &dtos.EncryptResponse{
		Data: encryptedItems, // adjust based on your encryption output
	}, nil
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
		encryptedItem[key] = prefix.AddKMSTypeIdentifier(encryptedVal)
	}

	return encryptedItem, nil
}

func (u *EncryptionUseCaseImpl) Decrypt(userID string, edekPrivate string, edekPublic string, req *dtos.DecryptRequest) (*dtos.DecryptResponse, error) {
	var DEKPrivate, DEKPublic []byte
	privateRetrieved := false
	publicRetrieved := false

	var decryptedItems []map[string]interface{}

	for _, item := range req.Data {
		decryptedItem := make(map[string]interface{})

		for key, value := range item {
			strVal := fmt.Sprintf("%v", value)

			var dek []byte

			// Determine DEK based on prefix
			if enums.HasKMSTypeIdentifier(strVal) {
				encryptedString, kmsType, _ := enums.RemoveKMSTypeIdentifier(strVal)

				switch *kmsType {
				case enums.KMSPrivate:
					if !privateRetrieved {
						kekID := enums.KMSPrivate.AddKMSTypeIdentifier(userID)
						kek, err := u.keyManager.RetrieveKEK(kekID)
						if err != nil {
							logger.Error("Failed to retrieve private KEK", err)
							return nil, errors.NewEncryptionError("failed to retrieve private KEK", err)
						}
						DEKPrivate, err = u.encryptionEngine.DecryptDEK(edekPrivate, kek)
						if err != nil {
							logger.Error("Failed to decrypt private DEK", err)
							return nil, errors.NewEncryptionError("failed to decrypt private DEK", err)
						}
						privateRetrieved = true
					}
					dek = DEKPrivate

				case enums.KMSPublic:
					if !publicRetrieved {
						kekID := enums.KMSPublic.AddKMSTypeIdentifier(userID)
						kek, err := u.keyManager.RetrieveKEK(kekID)
						if err != nil {
							logger.Error("Failed to retrieve public KEK", err)
							return nil, errors.NewEncryptionError("failed to retrieve public KEK", err)
						}
						DEKPublic, err = u.encryptionEngine.DecryptDEK(edekPublic, kek)
						if err != nil {
							logger.Error("Failed to decrypt public DEK", err)
							return nil, errors.NewEncryptionError("failed to decrypt public DEK", err)
						}
						publicRetrieved = true
					}
					dek = DEKPublic
				}

				plainText, err := u.encryptionEngine.Decrypt(encryptedString, dek)
				if err != nil {
					logger.Error("Failed to decrypt field", err)
					return nil, errors.NewEncryptionError(fmt.Sprintf("failed to decrypt field: %s", key), err)
				}
				decryptedItem[key] = plainText
			} else {
				logger.Error("Skipping field without recognized prefix", fmt.Errorf(key))
				decryptedItem[key] = strVal // leave unchanged or handle as error
			}
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
		kmsKeyName := keyType.AddKMSTypeIdentifier(userID)
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
	edekPrivate, err := generateEDEK(enums.KMSPrivate)
	if err != nil {
		return nil, err
	}

	edekPublic, err := generateEDEK(enums.KMSPublic)
	if err != nil {
		return nil, err
	}

	return &dtos.GenerateEDEKResponse{
		EDEKPrivate: edekPrivate,
		EDEKPublic:  edekPublic,
	}, nil
}
