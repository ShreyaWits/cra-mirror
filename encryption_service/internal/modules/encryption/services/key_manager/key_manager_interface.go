package keymanager

import (
	"context"
	"encryption_microservice/pkg/errors"
)

// KeyManagerInterface defines the interface for key management operations
type KeyManager interface {

	// StoreKEK stores a Key Encryption Key
	StoreKEK(ctx context.Context, keyID string, kek []byte) *errors.CustomError

	// RetrieveKEK retrieves a Key Encryption Key
	RetrieveKEK(ctx context.Context, keyID string) ([]byte, *errors.CustomError)
}
