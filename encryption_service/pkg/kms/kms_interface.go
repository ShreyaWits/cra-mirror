package kms

// KmsService defines the contract for key management operations using HashiCorp Vault
// as the underlying key management system. The interface supports role-based access
// control and uses Vault's transit secrets engine for key operations.
type KmsService interface {
	// GenerateKEK generates a new Key Encryption Key (KEK) using Vault's transit engine.
	// The key is generated with AES-256-GCM encryption and is suitable for both
	// end-to-end (ETE) and shared encryption scenarios.
	// Returns the generated key as a byte slice or an error if generation fails.
	GenerateKEK() ([]byte, error)

	// StoreKEK stores a Key Encryption Key (KEK) in Vault's transit engine.
	// Parameters:
	//   - kekID: Unique identifier for the key
	//   - kek: The key material to store (should be AES-256-GCM compatible)
	//   - role: The role for the key ("ete" for end-to-end or "shared" for shared access)
	// Returns an error if storage fails or if the role is invalid.
	StoreKEK(kekID string, kek []byte) error

	// RetrieveKEK retrieves a Key Encryption Key (KEK) from Vault's transit engine.
	// The key is retrieved only if the specified role matches the stored role.
	// Parameters:
	//   - kekID: Unique identifier for the key
	//   - role: The role to verify against ("ete" or "shared")
	// Returns the decrypted key material or an error if retrieval fails or role mismatch.
	RetrieveKEK(kekID string) ([]byte, error)

	// DeleteKEK deletes a Key Encryption Key (KEK) from Vault's transit engine.
	// The key is deleted only if the specified role matches the stored role.
	// Parameters:
	//   - kekID: Unique identifier for the key
	//   - role: The role to verify against ("ete" or "shared")
	// Returns an error if deletion fails or if the role doesn't match.
	DeleteKEK(kekID string) error

	// ListKEKs lists all Key Encryption Keys (KEKs) stored in Vault's transit engine
	// that match the specified role.
	// Parameters:
	//   - role: The role to filter keys by ("ete" or "shared")
	// Returns a slice of key IDs or an error if listing fails.
	ListKEKs() ([]string, error)
}

// Role constants for key management
const (
	// RoleETE represents end-to-end encryption keys
	RoleETE = "ete"
	// RoleShared represents shared encryption keys
	RoleShared = "shared"
)
