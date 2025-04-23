package kms

import (
	"encryption_microservice/pkg/errors"
	pkgErrors "encryption_microservice/pkg/errors"
	"fmt"
)

// TemporaryMockKMS is a temporary struct to implement the KMSInterface.
type TemporaryMockKMS struct {
	keks map[string][]byte // Mock storage for KEKs
}

// NewTemporaryMockKMS creates a new instance of TemporaryMockKMS.
func NewTemporaryMockKMS() KmsService {
	return &TemporaryMockKMS{
		keks: make(map[string][]byte),
	}
}

// GenerateKEK generates a new Key Encryption Key.
func (m *TemporaryMockKMS) GenerateKEK() ([]byte, *errors.CustomError) {
	// Mock implementation: return a static KEK
	return []byte("mock-generated-kek"), nil
}

// StoreKEK stores a Key Encryption Key in the KMS.
func (m *TemporaryMockKMS) StoreKEK(kekID string, kek []byte) *errors.CustomError {
	// Mock implementation: store the KEK in the map
	m.keks[kekID] = kek
	return nil
}

// RetrieveKEK retrieves a Key Encryption Key from the KMS.
func (m *TemporaryMockKMS) RetrieveKEK(kekID string) ([]byte, *errors.CustomError) {
	// Mock implementation: retrieve the KEK from the map
	kek, exists := m.keks[kekID]
	if !exists {
		return nil, pkgErrors.NewCustomError(pkgErrors.KMSerrRetrieveKEK, fmt.Errorf("KEK not found"))
	}
	return kek, nil
}

// DeleteKEK deletes a Key Encryption Key from the KMS.
func (m *TemporaryMockKMS) DeleteKEK(kekID string) *errors.CustomError {
	// Mock implementation: delete the KEK from the map
	delete(m.keks, kekID)
	return nil
}

// ListKEKs lists all Key Encryption Keys for a given role.
func (m *TemporaryMockKMS) ListKEKs() ([]string, *errors.CustomError) {
	// Mock implementation: return all KEK IDs
	var kekIDs []string
	for id := range m.keks {
		kekIDs = append(kekIDs, id)
	}
	return kekIDs, nil
}
