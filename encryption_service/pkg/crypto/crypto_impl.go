package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// Encryption provides AES-based encryption and decryption
type Encryption struct{}

// NewEncryption creates a new instance of Encryption
func NewEncryption() Encrypter {
	return &Encryption{}
}

// Encrypt encrypts a string using AES with CFB mode
func (e *Encryption) Encrypt(data string, secretKey []byte) (string, error) {

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	// Create an initialization vector (IV)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
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
func (e *Encryption) Decrypt(encryptedData string, secretKey []byte) (string, error) {
	// key := []byte(secretKey)

	// Decode the Base64-encoded ciphertext
	ciphertext, err := base64.URLEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	// Extract IV and ciphertext
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	// Decrypt the data
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

// EncryptBytes encrypts a byte slice using AES with CFB mode
func (e *Encryption) EncryptBytes(data []byte, secretKey []byte) ([]byte, error) {

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}

	// Create an initialization vector (IV)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Encrypt the data
	ciphertext := make([]byte, len(data))
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext, data)

	// Prepend IV to ciphertext
	return append(iv, ciphertext...), nil
}

// DecryptBytes decrypts an AES-encrypted byte slice in CFB mode
func (e *Encryption) DecryptBytes(encryptedData []byte, secretKey []byte) ([]byte, error) {

	if len(encryptedData) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	// Extract IV and ciphertext
	iv := encryptedData[:aes.BlockSize]
	ciphertext := encryptedData[aes.BlockSize:]

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}

	// Decrypt the data
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}
