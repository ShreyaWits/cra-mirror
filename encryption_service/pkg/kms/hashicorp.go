package kms

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"encryption_microservice/internal/config"
	"encryption_microservice/pkg/errors"
	"encryption_microservice/pkg/logger"

	"github.com/hashicorp/vault/api"
)

// HashiCorpKMS implements the KmsService interface using HashiCorp Vault
type HashiCorpKMS struct {
	client      *api.Client
	transitPath string
}

// NewHashiCorpKMS creates a new instance of HashiCorpKMS
func NewHashiCorpKMS(cfg *config.Config) KmsService {
	// Initialize Vault client
	vaultConfig := api.DefaultConfig()
	vaultConfig.Address = cfg.VaultAddr
	vaultConfig.MaxRetries = 3
	vaultConfig.Timeout = 10 * time.Second

	vaultClient, err := api.NewClient(vaultConfig)
	if err != nil {
		log.Fatalf("Failed to create Vault client: %v", err)
	}

	// Set Vault token
	vaultClient.SetToken(cfg.VaultToken)
	return &HashiCorpKMS{
		client:      vaultClient,
		transitPath: cfg.VaultPath,
	}
}

// GenerateKEK generates a new Key Encryption Key using Vault's transit engine
func (h *HashiCorpKMS) GenerateKEK() ([]byte, *errors.CustomError) {
	// Generate a new key using Vault's transit engine
	path := filepath.Join(h.transitPath, "keys", "kek")
	data := map[string]interface{}{
		"type": "aes256-gcm96",
	}

	_, err := h.client.Logical().WriteWithContext(context.Background(), path, data)
	if err != nil {
		return nil, errors.NewCustomError(errors.KMSerrGenerateKEK, err)
	}

	// Generate a random key using Vault's transit engine
	path = filepath.Join(h.transitPath, "datakey", "plaintext", "kek")
	secret, err := h.client.Logical().WriteWithContext(context.Background(), path, nil)
	if err != nil {
		return nil, errors.NewCustomError(errors.KMSerrRandomKey, err)
	}

	plaintext, ok := secret.Data["plaintext"].(string)
	if !ok {
		return nil, errors.NewCustomError(errors.KMSerrRandomKey, fmt.Errorf("invalid response format from Vault"))
	}

	key, err := base64.StdEncoding.DecodeString(plaintext)
	if err != nil {
		return nil, errors.NewCustomError(errors.KMSerrRandomKey, err)
	}

	return key, nil
}

// StoreKEK stores a Key Encryption Key in Vault's KV engine
func (h *HashiCorpKMS) StoreKEK(kekID string, kek []byte) *errors.CustomError {
	keyBase64 := base64.StdEncoding.EncodeToString(kek)
	path := filepath.Join("secret/data", kekID)
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"key": keyBase64,
		},
	}
	_, err := h.client.Logical().WriteWithContext(context.Background(), path, data)
	if err != nil {
		return errors.NewCustomError(errors.KMSerrStoreKEK, err)
	}
	return nil
}

// RetrieveKEK retrieves a Key Encryption Key from Vault's KV engine
func (h *HashiCorpKMS) RetrieveKEK(kekID string) ([]byte, *errors.CustomError) {
	path := filepath.Join("secret/data", kekID)
	secret, err := h.client.Logical().ReadWithContext(context.Background(), path)
	if err != nil {
		return nil, errors.NewCustomError(errors.KMSerrRetrieveKEK, err)
	}

	if secret == nil || secret.Data == nil {
		return nil, errors.NewCustomError(errors.KMSerrRetrieveKEK, fmt.Errorf("no KEK found at path: %s", path))
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, errors.NewCustomError(errors.KMSerrRetrieveKEK, fmt.Errorf("unexpected secret data structure"))
	}

	plaintext, ok := data["key"].(string)
	if !ok {
		return nil, errors.NewCustomError(errors.KMSerrRetrieveKEK, fmt.Errorf("missing key field in Vault response"))
	}

	decoded, err := base64.StdEncoding.DecodeString(plaintext)
	if err != nil {
		return nil, errors.NewCustomError(errors.KMSerrRetrieveKEK, err)
	}

	return decoded, nil
}

// DeleteKEK deletes a Key Encryption Key from Vault
func (h *HashiCorpKMS) DeleteKEK(kekID string) *errors.CustomError {
	// Delete the key from Vault's transit engine
	path := filepath.Join(h.transitPath, "keys", kekID)
	_, err := h.client.Logical().DeleteWithContext(context.Background(), path)
	if err != nil {
		return errors.NewCustomError(errors.KMSerrDeleteKEK, err)
	}

	return nil
}

// ListKEKs lists all Key Encryption Keys
func (h *HashiCorpKMS) ListKEKs() ([]string, *errors.CustomError) {
	// List all keys in the transit engine
	path := filepath.Join(h.transitPath, "keys")
	secret, err := h.client.Logical().ListWithContext(context.Background(), path)
	if err != nil {
		return nil, errors.NewCustomError(errors.KMSerrListKEKs, err)
	}

	var keys []string
	if secret != nil && secret.Data != nil {
		if keyList, ok := secret.Data["keys"].([]interface{}); ok {
			for _, key := range keyList {
				keys = append(keys, key.(string))
			}
		}
	}

	return keys, nil
}

// Encrypt encrypts data using a Key Encryption Key
func (h *HashiCorpKMS) Encrypt(kekID string, data []byte) ([]byte, *errors.CustomError) {
	path := fmt.Sprintf("%s/encrypt/%s", h.transitPath, kekID)
	requestData := map[string]interface{}{
		"plaintext": base64.StdEncoding.EncodeToString(data),
	}

	secret, err := h.client.Logical().Write(path, requestData)
	if err != nil {
		logger.Error("Failed to encrypt data with Vault", err)
		return nil, errors.NewCustomError(errors.KMSerrEncryptData, err)
	}

	ciphertext, ok := secret.Data["ciphertext"].(string)
	if !ok {
		return nil, errors.NewCustomError(errors.KMSerrEncryptData, fmt.Errorf("invalid ciphertext format from Vault"))
	}

	return []byte(ciphertext), nil
}

// Decrypt decrypts data using a Key Encryption Key
func (h *HashiCorpKMS) Decrypt(kekID string, ciphertext []byte) ([]byte, *errors.CustomError) {
	path := fmt.Sprintf("%s/decrypt/%s", h.transitPath, kekID)
	requestData := map[string]interface{}{
		"ciphertext": string(ciphertext),
	}

	secret, err := h.client.Logical().Write(path, requestData)
	if err != nil {
		logger.Error("Failed to decrypt data with Vault", err)
		return nil, errors.NewCustomError(errors.KMSerrDecryptData, err)
	}

	plaintext, ok := secret.Data["plaintext"].(string)
	if !ok {
		return nil, errors.NewCustomError(errors.KMSerrDecryptData, fmt.Errorf("invalid plaintext format from Vault"))
	}

	decoded, err := base64.StdEncoding.DecodeString(plaintext)
	if err != nil {
		return nil, errors.NewCustomError(errors.KMSerrDecryptData, err)
	}

	return decoded, nil
}
