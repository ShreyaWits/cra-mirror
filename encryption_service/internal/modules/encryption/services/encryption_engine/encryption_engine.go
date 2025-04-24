package encryptionengine

import (
	"crypto/rand"
	"encoding/base64"
	"io"

	keymanager "encryption_microservice/internal/modules/encryption/services/key_manager"
	"encryption_microservice/pkg/crypto"
	"encryption_microservice/pkg/errors"
)

type EncryptionEngineImpl struct {
	keyManager keymanager.KeyManager
	crypto     crypto.Encrypter
}

func NewEncryptionEngine(keyManager keymanager.KeyManager, crypto crypto.Encrypter) EncryptionEngine {
	return &EncryptionEngineImpl{
		keyManager: keyManager,
		crypto:     crypto,
	}
}

func (e *EncryptionEngineImpl) GenerateEncryptionKey() ([]byte, *errors.CustomError) {
	dek := make([]byte, 32) // 256 bits for AES-*errors.CustomError
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, errors.NewCustomError(errors.ENGErrGenerateDEK, err)
	}
	return dek, nil
}

func (e *EncryptionEngineImpl) Encrypt(data string, dek []byte) (string, *errors.CustomError) {

	// Encrypt the DEK with KEK
	encryptedData, err := e.crypto.Encrypt(data, dek)
	if err != nil {
		return "", err
	}

	return encryptedData, nil
}

func (e *EncryptionEngineImpl) Decrypt(encryptedData string, dek []byte) (string, *errors.CustomError) {
	// Encrypt the DEK with KEK
	decryptedData, err := e.crypto.Decrypt(encryptedData, dek)
	if err != nil {
		return "", err
	}

	return decryptedData, nil
}

func (e *EncryptionEngineImpl) EncryptDEK(dek []byte, kek []byte) (string, *errors.CustomError) {
	// Create AES cipher with KEK
	edek, err := e.crypto.EncryptBytes(dek, kek)
	if err != nil {
		return "", err
	}
	// Encode the encrypted DEK
	return base64.StdEncoding.EncodeToString(edek), nil
}

func (e *EncryptionEngineImpl) DecryptDEK(edek string, kek []byte) ([]byte, *errors.CustomError) {
	// Encrypt the DEK with KEK
	decryptedEDEK, err :=
		base64.StdEncoding.DecodeString(edek)
	if err != nil {
		return nil, errors.NewCustomError(errors.ENGErrDecryptDEK, err)
	}
	if decryptedData, err := e.crypto.DecryptBytes(decryptedEDEK, kek); err != nil {
		return nil, errors.NewCustomError(errors.ENGErrDecryptDEK, err)
	} else {
		return decryptedData, nil
	}
}
