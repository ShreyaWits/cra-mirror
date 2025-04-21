package crypto

// Encrypter defines methods for encryption and decryption
type Encrypter interface {
	Encrypt(data string, secretKey string) (string, error)
	Decrypt(encryptedData string, secretKey string) (string, error)
	EncryptBytes(data []byte, secretKey string) ([]byte, error)
	DecryptBytes(encryptedData []byte, secretKey string) ([]byte, error)
}
