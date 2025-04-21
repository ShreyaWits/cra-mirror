package keymanager

// KeyManagerInterface defines the interface for key management operations
type KeyManager interface {
	// GenerateKEK generates a new Key Encryption Key
	GenerateKEK() ([]byte, error)

	// StoreKEK stores a Key Encryption Key
	StoreKEK(keyID string, kek []byte) error

	// RetrieveKEK retrieves a Key Encryption Key
	RetrieveKEK(keyID string) ([]byte, error)

	// DeleteKEK deletes a Key Encryption Key
	DeleteKEK(keyID string) error

	// ListKEKs lists all Key Encryption Keys
	ListKEKs() ([]string, error)
}
