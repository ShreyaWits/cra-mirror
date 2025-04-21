package keymanager

import (
	"crypto/rand"
	"encryption_microservice/pkg/errors"
	"encryption_microservice/pkg/kms"
	"io"
)

type KeyManagerImpl struct {
	kms kms.KmsService
}

func NewKeyManager(kms kms.KmsService) KeyManager {
	return &KeyManagerImpl{
		kms: kms,
	}
}

func (k *KeyManagerImpl) GenerateKEK() ([]byte, error) {
	kek := make([]byte, 32) // 256 bits for AES-256
	if _, err := io.ReadFull(rand.Reader, kek); err != nil {
		return nil, errors.NewEncryptionError("failed to generate KEK", err)
	}
	return kek, nil
}

func (k *KeyManagerImpl) StoreKEK(keyID string, kek []byte) error {
	if err := k.kms.StoreKEK(keyID, kek); err != nil {
		return errors.NewEncryptionError("failed to store KEK", err)
	}
	return nil
}

func (k *KeyManagerImpl) RetrieveKEK(keyID string) ([]byte, error) {
	kek, err := k.kms.RetrieveKEK(keyID)
	if err != nil {
		return nil, errors.NewEncryptionError("failed to retrieve KEK", err)
	}
	return kek, nil
}

func (k *KeyManagerImpl) DeleteKEK(keyID string) error {
	if err := k.kms.DeleteKEK(keyID); err != nil {
		return errors.NewEncryptionError("failed to delete KEK", err)
	}
	return nil
}

func (k *KeyManagerImpl) ListKEKs() ([]string, error) {
	keyIDs, err := k.kms.ListKEKs()
	if err != nil {
		return nil, errors.NewEncryptionError("failed to list KEKs", err)
	}
	return keyIDs, nil
}
