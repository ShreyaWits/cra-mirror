package crypto

// Encrypter defines methods for encryption and decryption
type Encrypter interface {
	Encrypt(data string, secretKey []byte) (string, error)
	Decrypt(encryptedData string, secretKey []byte) (string, error)
	EncryptBytes(data []byte, secretKey []byte) ([]byte, error)
	DecryptBytes(encryptedData []byte, secretKey []byte) ([]byte, error)
}
