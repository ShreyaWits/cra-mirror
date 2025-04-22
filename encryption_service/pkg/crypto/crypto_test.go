package crypto

// import (
// 	"testing"
// )

// // TestEncryption tests AES encryption and decryption methods
// func TestEncryption(t *testing.T) {
// 	secretKey := "mysecretkey12345" // 16 bytes for AES-128
// 	data := "This is a secret message"

// 	// Create a new instance of Encryption
// 	encrypter := NewEncryption()

// 	// Test Encrypt and Decrypt with string data
// 	encryptedData, err := encrypter.Encrypt(data, secretKey)
// 	if err != nil {
// 		t.Fatalf("Encryption failed: %v", err)
// 	}
// 	t.Logf("Encrypted string: %s", encryptedData)

// 	decryptedData, err := encrypter.Decrypt(encryptedData, secretKey)
// 	if err != nil {
// 		t.Fatalf("Decryption failed: %v", err)
// 	}
// 	t.Logf("Decrypted string: %s", decryptedData)

// 	if decryptedData != data {
// 		t.Errorf("Expected decrypted data to be %s, but got %s", data, decryptedData)
// 	}

// 	// Test EncryptBytes and DecryptBytes with byte data
// 	encryptedBytes, err := encrypter.EncryptBytes([]byte(data), secretKey)
// 	if err != nil {
// 		t.Fatalf("EncryptBytes failed: %v", err)
// 	}
// 	t.Logf("Encrypted bytes: %v", encryptedBytes)

// 	decryptedBytes, err := encrypter.DecryptBytes(encryptedBytes, secretKey)
// 	if err != nil {
// 		t.Fatalf("DecryptBytes failed: %v", err)
// 	}
// 	t.Logf("Decrypted bytes: %s", decryptedBytes)

// 	if string(decryptedBytes) != data {
// 		t.Errorf("Expected decrypted bytes to be %s, but got %s", data, decryptedBytes)
// 	}
// }

// // TestEncryptionWithInvalidKey tests encryption and decryption with an invalid key
// func TestEncryptionWithInvalidKey(t *testing.T) {
// 	secretKey := "mysecretkey12345" // 16 bytes for AES-128
// 	invalidKey := "invalidkey12345" // 16 bytes for AES-128, but incorrect value
// 	data := "This is a secret message"

// 	// Create a new instance of Encryption
// 	encrypter := NewEncryption()

// 	// Encrypt with valid key
// 	encryptedData, err := encrypter.Encrypt(data, secretKey)
// 	if err != nil {
// 		t.Fatalf("Encryption failed: %v", err)
// 	}
// 	t.Logf("Encrypted data with valid key: %s", encryptedData)

// 	// Attempt to decrypt with an invalid key
// 	_, err = encrypter.Decrypt(encryptedData, invalidKey)
// 	if err == nil {
// 		t.Fatalf("Expected error when decrypting with an invalid key, but got nil")
// 	} else {
// 		t.Logf("Error as expected when decrypting with invalid key: %v", err)
// 	}
// }

// // TestDecryptionWithInvalidData tests decryption with invalid data (wrong ciphertext format)
// func TestDecryptionWithInvalidData(t *testing.T) {
// 	secretKey := "mysecretkey12345" // 16 bytes for AES-128
// 	invalidData := "invalidEncryptedData"

// 	// Create a new instance of Encryption
// 	encrypter := NewEncryption()

// 	// Attempt to decrypt invalid data
// 	_, err := encrypter.Decrypt(invalidData, secretKey)
// 	if err == nil {
// 		t.Fatalf("Expected error when decrypting invalid data, but got nil")
// 	} else {
// 		t.Logf("Error as expected when decrypting invalid data: %v", err)
// 	}
// }

// // TestEncryptionWithShortData tests encryption with very short data (e.g., 1 byte)
// func TestEncryptionWithShortData(t *testing.T) {
// 	secretKey := "mysecretkey12345" // 16 bytes for AES-128
// 	data := "A"                     // 1 byte of data

// 	// Create a new instance of Encryption
// 	encrypter := NewEncryption()

// 	// Test Encrypt and Decrypt with short data
// 	encryptedData, err := encrypter.Encrypt(data, secretKey)
// 	if err != nil {
// 		t.Fatalf("Encryption failed: %v", err)
// 	}
// 	t.Logf("Encrypted data with short string: %s", encryptedData)

// 	decryptedData, err := encrypter.Decrypt(encryptedData, secretKey)
// 	if err != nil {
// 		t.Fatalf("Decryption failed: %v", err)
// 	}
// 	t.Logf("Decrypted data with short string: %s", decryptedData)

// 	if decryptedData != data {
// 		t.Errorf("Expected decrypted data to be %s, but got %s", data, decryptedData)
// 	}
// }
