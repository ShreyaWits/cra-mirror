package kms

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"encryption_microservice/internal/config"
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
func (h *HashiCorpKMS) GenerateKEK() ([]byte, error) {
	// Generate a new key using Vault's transit engine
	path := filepath.Join(h.transitPath, "keys", "kek")
	data := map[string]interface{}{
		"type": "aes256-gcm96",
	}

	_, err := h.client.Logical().WriteWithContext(context.Background(), path, data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate KEK: %w", err)
	}

	// Generate a random key using Vault's transit engine
	path = filepath.Join(h.transitPath, "datakey", "plaintext", "kek")
	secret, err := h.client.Logical().WriteWithContext(context.Background(), path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	// Extract the plaintext key from the response
	plaintext, ok := secret.Data["plaintext"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid response format from Vault")
	}

	// Decode the base64 encoded key
	key, err := base64.StdEncoding.DecodeString(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key: %w", err)
	}

	return key, nil
}

// StoreKEK stores a Key Encryption Key in Vault's KV engine
func (h *HashiCorpKMS) StoreKEK(kekID string, kek []byte) error {
	keyBase64 := base64.StdEncoding.EncodeToString(kek)
	path := filepath.Join("secret/data", kekID)
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"key": keyBase64,
		},
	}
	_, err := h.client.Logical().WriteWithContext(context.Background(), path, data)
	if err != nil {
		return fmt.Errorf("failed to store KEK: %w", err)
	}
	return nil
}

// RetrieveKEK retrieves a Key Encryption Key from Vault's KV engine
func (h *HashiCorpKMS) RetrieveKEK(kekID string) ([]byte, error) {
	path := filepath.Join("secret/data", kekID)
	secret, err := h.client.Logical().ReadWithContext(context.Background(), path)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve KEK: %w", err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no KEK found at path: %s", path)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected secret data structure")
	}

	plaintext, ok := data["key"].(string)
	if !ok {
		return nil, fmt.Errorf("missing key field in Vault response")
	}

	return base64.StdEncoding.DecodeString(plaintext)
}

// DeleteKEK deletes a Key Encryption Key from Vault
func (h *HashiCorpKMS) DeleteKEK(kekID string) error {
	// Delete the key from Vault's transit engine
	path := filepath.Join(h.transitPath, "keys", kekID)
	_, err := h.client.Logical().DeleteWithContext(context.Background(), path)
	if err != nil {
		return fmt.Errorf("failed to delete KEK: %w", err)
	}

	return nil
}

// ListKEKs lists all Key Encryption Keys
func (h *HashiCorpKMS) ListKEKs() ([]string, error) {
	// List all keys in the transit engine
	path := filepath.Join(h.transitPath, "keys")
	secret, err := h.client.Logical().ListWithContext(context.Background(), path)
	if err != nil {
		return nil, fmt.Errorf("failed to list KEKs: %w", err)
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
func (h *HashiCorpKMS) Encrypt(kekID string, data []byte) ([]byte, error) {
	path := fmt.Sprintf("%s/encrypt/%s", h.transitPath, kekID)
	requestData := map[string]interface{}{
		"plaintext": base64.StdEncoding.EncodeToString(data),
	}

	secret, err := h.client.Logical().Write(path, requestData)
	if err != nil {
		logger.Error("Failed to encrypt data with Vault", err)
		return nil, fmt.Errorf("failed to encrypt data with Vault: %w", err)
	}

	ciphertext, ok := secret.Data["ciphertext"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid ciphertext format from Vault")
	}

	return []byte(ciphertext), nil
}

// Decrypt decrypts data using a Key Encryption Key
func (h *HashiCorpKMS) Decrypt(kekID string, ciphertext []byte) ([]byte, error) {
	path := fmt.Sprintf("%s/decrypt/%s", h.transitPath, kekID)
	requestData := map[string]interface{}{
		"ciphertext": string(ciphertext),
	}

	secret, err := h.client.Logical().Write(path, requestData)
	if err != nil {
		logger.Error("Failed to decrypt data with Vault", err)
		return nil, fmt.Errorf("failed to decrypt data with Vault: %w", err)
	}

	plaintext, ok := secret.Data["plaintext"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid plaintext format from Vault")
	}

	decoded, err := base64.StdEncoding.DecodeString(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode plaintext: %w", err)
	}

	return decoded, nil
}
