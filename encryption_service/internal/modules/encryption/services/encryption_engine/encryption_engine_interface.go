package encryptionengine

import (
	"context"
	"encryption_microservice/pkg/errors"
)

// EncryptionEngine defines the interface for encryption operations
type EncryptionEngine interface {

	// GenerateDEK generates a new Encryption Key
	GenerateEncryptionKey(ctx context.Context) ([]byte, *errors.CustomError)

	// Encrypt encrypts data and returns the encrypted data and EDEK
	Encrypt(ctx context.Context, data string, dek []byte) (string, *errors.CustomError)

	// Decrypt decrypts data using the EDEK and returns the original data
	Decrypt(ctx context.Context, encryptedData string, dek []byte) (string, *errors.CustomError)

	// EncryptDEK encrypts a Data Encryption Key with a Key Encryption Key
	EncryptDEK(ctx context.Context, dek []byte, kek []byte) (string, *errors.CustomError)

	// DecryptDEK decrypts an Encrypted Data Encryption Key with a Key Encryption Key
	DecryptDEK(ctx context.Context, edek string, kek []byte) ([]byte, *errors.CustomError)
}
