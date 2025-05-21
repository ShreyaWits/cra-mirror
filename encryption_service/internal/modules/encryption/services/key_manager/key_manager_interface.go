package keymanager

import "encryption_microservice/pkg/errors"

// KeyManagerInterface defines the interface for key management operations
type KeyManager interface {

	// StoreKEK stores a Key Encryption Key
	StoreKEK(keyID string, kek []byte) *errors.CustomError

	// RetrieveKEK retrieves a Key Encryption Key
	RetrieveKEK(keyID string) ([]byte, *errors.CustomError)
}
