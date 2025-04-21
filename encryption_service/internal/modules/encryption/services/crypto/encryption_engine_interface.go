package crypto

// EncryptionEngine defines the interface for encryption operations
type EncryptionEngine interface {

	// GenerateDEK generates a new Encryption Key
	GenerateEncryptionKey() ([]byte, error)

	// Encrypt encrypts data and returns the encrypted data and EDEK
	Encrypt(data string, dek []byte) (string, error)

	// Decrypt decrypts data using the EDEK and returns the original data
	Decrypt(encryptedData string, dek []byte) (string, error)

	// EncryptDEK encrypts a Data Encryption Key with a Key Encryption Key
	EncryptDEK(dek []byte, kek []byte) (string, error)

	// DecryptDEK decrypts an Encrypted Data Encryption Key with a Key Encryption Key
	DecryptDEK(edek string, kek []byte) ([]byte, error)
}
