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

func (k *KeyManagerImpl) GenerateKEK() ([]byte, *errors.CustomError) {
	kek := make([]byte, 32) // 256 bits for AES-256
	if _, err := io.ReadFull(rand.Reader, kek); err != nil {
		return nil, errors.NewCustomError(errors.KMGErrGenrateKEK, err)
	}
	return kek, nil
}

func (k *KeyManagerImpl) StoreKEK(keyID string, kek []byte) *errors.CustomError {
	if err := k.kms.StoreKEK(keyID, kek); err != nil {
		return err
	}
	return nil
}

func (k *KeyManagerImpl) RetrieveKEK(keyID string) ([]byte, *errors.CustomError) {
	kek, err := k.kms.RetrieveKEK(keyID)
	if err != nil {
		return nil, err
	}
	return kek, nil
}

func (k *KeyManagerImpl) DeleteKEK(keyID string) *errors.CustomError {
	if err := k.kms.DeleteKEK(keyID); err != nil {
		return err
	}
	return nil
}

func (k *KeyManagerImpl) ListKEKs() ([]string, *errors.CustomError) {
	keyIDs, err := k.kms.ListKEKs()
	if err != nil {
		return nil, err
	}
	return keyIDs, nil
}
