package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"

	keymanager "encryption_microservice/internal/modules/encryption/services/key_manager"
	"encryption_microservice/pkg/errors"
)

type EncryptionEngineImpl struct {
	keyManager keymanager.KeyManager
}

func NewEncryptionEngine(keyManager keymanager.KeyManager) EncryptionEngine {
	return &EncryptionEngineImpl{
		keyManager: keyManager,
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

	// Create AES cipher
	block, err := aes.NewCipher(dek)
	if err != nil {
		return "", errors.NewEncryptionError("failed to create cipher", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.NewEncryptionError("failed to create GCM", err)
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.NewEncryptionError("failed to generate nonce", err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, []byte(data), []byte("contextID"))

	// Encode the encrypted data
	encryptedData := base64.StdEncoding.EncodeToString(ciphertext)

	// Get KEK from key manager
	kek, err := e.keyManager.RetrieveKEK("contextID")
	if err != nil {
		return "", err
	}

	// Encrypt the DEK with KEK
	_, err = e.EncryptDEK(dek, kek)
	if err != nil {
		return "", err
	}

	return encryptedData, nil
}

func (e *EncryptionEngineImpl) Decrypt(encryptedData string, dek []byte) (string, error) {
	// Decode the encrypted data
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", errors.NewEncryptionError("failed to decode encrypted data", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(dek)
	if err != nil {
		return "", errors.NewEncryptionError("failed to create cipher", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.NewEncryptionError("failed to create GCM", err)
	}

	// Extract nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.NewEncryptionError("ciphertext too short", nil)
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the data using the same context ID
	plaintext, err := gcm.Open(nil, nonce, ciphertext, []byte("contextID"))
	if err != nil {
		return "", errors.NewEncryptionError("failed to decrypt data", err)
	}

	return string(plaintext), nil
}

func (e *EncryptionEngineImpl) EncryptDEK(dek []byte, kek []byte) (string, error) {
	// Create AES cipher with KEK
	block, err := aes.NewCipher(kek)
	if err != nil {
		return "", errors.NewEncryptionError("failed to create cipher for DEK encryption", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.NewEncryptionError("failed to create GCM for DEK encryption", err)
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.NewEncryptionError("failed to generate nonce for DEK encryption", err)
	}

	// Encrypt the DEK
	ciphertext := gcm.Seal(nonce, nonce, dek, nil)

	// Encode the encrypted DEK
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *EncryptionEngineImpl) DecryptDEK(edek string, kek []byte) ([]byte, error) {
	// Decode the encrypted DEK (EDEK) from base64
	ciphertext, err := base64.StdEncoding.DecodeString(edek)
	if err != nil {
		return nil, errors.NewEncryptionError("failed to decode encrypted DEK", err)
	}

	// Ensure the KEK is of valid size for AES (16, 24, or 32 bytes)
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, errors.NewEncryptionError("failed to create AES cipher for DEK decryption", err)
	}

	// Create GCM mode of operation
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.NewEncryptionError("failed to create GCM mode for DEK decryption", err)
	}

	// Extract nonce from the ciphertext (first part)
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.NewEncryptionError("ciphertext is too short for valid nonce extraction", nil)
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the ciphertext using the GCM
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.NewEncryptionError("failed to decrypt DEK, authentication failed", err)
	}

	return plaintext, nil
}
