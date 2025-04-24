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

// NewHashiCorpKMSWithClient creates a HashiCorpKMS using the provided Vault client and transit path, for testing purposes.
func NewHashiCorpKMSWithClient(client *api.Client, transitPath string) KmsService {
	return &HashiCorpKMS{client: client, transitPath: transitPath}
}

// StoreKEK stores a Key Encryption Key in Vault's KV engine
func (h *HashiCorpKMS) StoreKEK(kekID string, kek []byte) *errors.CustomError {
	keyBase64 := base64.StdEncoding.EncodeToString(kek)
	path := filepath.Join("secret/data", kekID)
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"key": keyBase64,
		},
		"deletion_allowed": true,
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
