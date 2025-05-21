package crypto

import "encryption_microservice/pkg/errors"

// Encrypter defines methods for encryption and decryption
type Encrypter interface {
	Encrypt(data string, secretKey []byte) (string, *errors.CustomError)
	Decrypt(encryptedData string, secretKey []byte) (string, *errors.CustomError)
	EncryptBytes(data []byte, secretKey []byte) ([]byte, *errors.CustomError)
	DecryptBytes(encryptedData []byte, secretKey []byte) ([]byte, *errors.CustomError)
}
