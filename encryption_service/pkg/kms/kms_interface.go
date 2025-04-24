package kms

import "encryption_microservice/pkg/errors"

// KmsService defines the contract for key management operations using HashiCorp Vault
// as the underlying key management system. The interface supports role-based access
// control and uses Vault's transit secrets engine for key operations.
type KmsService interface {

	// StoreKEK stores a Key Encryption Key (KEK) in Vault's transit engine.
	// Parameters:
	//   - kekID: Unique identifier for the key
	//   - kek: The key material to store (should be AES-256-GCM compatible)
	//   - role: The role for the key ("ete" for end-to-end or "shared" for shared access)
	// Returns an error if storage fails or if the role is invalid.
	StoreKEK(kekID string, kek []byte) *errors.CustomError

	// RetrieveKEK retrieves a Key Encryption Key (KEK) from Vault's transit engine.
	RetrieveKEK(kekID string) ([]byte, *errors.CustomError)
}
