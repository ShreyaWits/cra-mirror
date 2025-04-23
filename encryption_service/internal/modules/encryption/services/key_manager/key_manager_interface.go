package keymanager

import "encryption_microservice/pkg/errors"

// KeyManagerInterface defines the interface for key management operations
type KeyManager interface {
	// GenerateKEK generates a new Key Encryption Key
	GenerateKEK() ([]byte, *errors.CustomError)

	// StoreKEK stores a Key Encryption Key
	StoreKEK(keyID string, kek []byte) *errors.CustomError

	// RetrieveKEK retrieves a Key Encryption Key
	RetrieveKEK(keyID string) ([]byte, *errors.CustomError)

	// DeleteKEK deletes a Key Encryption Key
	DeleteKEK(keyID string) *errors.CustomError

	// ListKEKs lists all Key Encryption Keys
	ListKEKs() ([]string, *errors.CustomError)
}
