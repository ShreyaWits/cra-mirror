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

func (e *EncryptionEngineImpl) GenerateEncryptionKey() ([]byte, error) {
	dek := make([]byte, 32) // 256 bits for AES-256
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, errors.NewEncryptionError("failed to generate DEK", err)
	}
	return dek, nil
}

func (e *EncryptionEngineImpl) Encrypt(data string, dek []byte) (string, error) {

	// Encrypt the DEK with KEK
	encryptedData, err := e.crypto.Encrypt(data, string(dek))
	if err != nil {
		return "", err
	}

	return encryptedData, nil
}

func (e *EncryptionEngineImpl) Decrypt(encryptedData string, dek []byte) (string, error) {
	// Encrypt the DEK with KEK
	decryptedData, err := e.crypto.Decrypt(encryptedData, string(dek))
	if err != nil {
		return "", err
	}

	return decryptedData, nil
}

func (e *EncryptionEngineImpl) EncryptDEK(dek []byte, kek []byte) (string, error) {
	// Create AES cipher with KEK
	edek, err := e.crypto.EncryptBytes(dek, string(kek))
	if err != nil {
		return "", err
	}
	// Encode the encrypted DEK
	return base64.StdEncoding.EncodeToString(edek), nil
}

func (e *EncryptionEngineImpl) DecryptDEK(edek string, kek []byte) ([]byte, error) {
	// Encrypt the DEK with KEK
	decryptedEDEK, err :=
		base64.StdEncoding.DecodeString(edek)
	if err != nil {
		return nil, err
	}
	decryptedData, err := e.crypto.DecryptBytes(decryptedEDEK, string(kek))
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}
