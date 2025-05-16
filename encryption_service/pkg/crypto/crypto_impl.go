package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	pkgErrors "encryption_microservice/pkg/errors"
	"fmt"
	"io"
)

// Encryption provides AES-based encryption and decryption
type Encryption struct{}

// NewEncryption creates a new instance of Encryption
func NewEncryption() Encrypter {
	return &Encryption{}
}

// Encrypt encrypts a string using AES with CFB mode
func (e *Encryption) Encrypt(data string, secretKey []byte) (string, *pkgErrors.CustomError) {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", pkgErrors.NewCustomError(pkgErrors.CRYPerrEncryptData, err)
	}

	// Create an initialization vector (IV)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", pkgErrors.NewCustomError(pkgErrors.CRYPerrEncryptData, err)
	}

	// Encrypt the data
	ciphertext := make([]byte, len(data))
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext, []byte(data))

	// Prepend IV to ciphertext
	finalCiphertext := append(iv, ciphertext...)

	// Return Base64-encoded result
	return base64.URLEncoding.EncodeToString(finalCiphertext), nil
}

// Decrypt decrypts an AES-encrypted string in CFB mode
func (e *Encryption) Decrypt(encryptedData string, secretKey []byte) (string, *pkgErrors.CustomError) {
	// key := []byte(secretKey)

	// Decode the Base64-encoded ciphertext
	ciphertext, err := base64.URLEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", pkgErrors.NewCustomError(pkgErrors.CRYPerrDecryptData, err)
	}

	if len(ciphertext) < aes.BlockSize {
		return "", pkgErrors.NewCustomError(pkgErrors.CRYPerrDecryptData, fmt.Errorf("ciphertext too short"))
	}

	// Extract IV and ciphertext
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", pkgErrors.NewCustomError(pkgErrors.CRYPerrDecryptData, err)
	}

	// Decrypt the data
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

// EncryptBytes encrypts a byte slice using AES with CFB mode
func (e *Encryption) EncryptBytes(data []byte, secretKey []byte) ([]byte, *pkgErrors.CustomError) {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, pkgErrors.NewCustomError(pkgErrors.CRYPerrEncryptBytes, err)
	}

	// Create an initialization vector (IV)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, pkgErrors.NewCustomError(pkgErrors.CRYPerrEncryptBytes, err)
	}

	// Encrypt the data
	ciphertext := make([]byte, len(data))
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext, data)

	// Prepend IV to ciphertext
	return append(iv, ciphertext...), nil
}

// DecryptBytes decrypts an AES-encrypted byte slice in CFB mode
func (e *Encryption) DecryptBytes(encryptedData []byte, secretKey []byte) ([]byte, *pkgErrors.CustomError) {
	if len(encryptedData) < aes.BlockSize {
		return nil, pkgErrors.NewCustomError(pkgErrors.CRYPerrDecryptBytes, fmt.Errorf("ciphertext too short"))
	}

	// Extract IV and ciphertext
	iv := encryptedData[:aes.BlockSize]
	ciphertext := encryptedData[aes.BlockSize:]

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, pkgErrors.NewCustomError(pkgErrors.CRYPerrDecryptBytes, err)
	}

	// Decrypt the data
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}
