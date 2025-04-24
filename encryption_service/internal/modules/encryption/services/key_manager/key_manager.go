package keymanager

import (
	"encryption_microservice/pkg/errors"
	"encryption_microservice/pkg/kms"
)

type KeyManagerImpl struct {
	kms kms.KmsService
}

func NewKeyManager(kms kms.KmsService) KeyManager {
	return &KeyManagerImpl{
		kms: kms,
	}
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
