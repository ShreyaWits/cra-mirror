package kms

import (
	"testing"
	"time"

	"encryption_microservice/internal/config"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//NOTE :- RUN `vault server -dev` command and copy ROOT Token in SetToken

func setupTestVault(t *testing.T) (*api.Client, string) {
	// Create a Vault client
	config := api.DefaultConfig()
	config.Address = "http://localhost:8200" // Default Vault dev server address

	client, err := api.NewClient(config)
	require.Nil(t, err, "Failed to create Vault client")

	// Set the root token (for dev server)
	client.SetToken("hvs.MlbuTDJ6DqdtgnFNuiOmNKej")

	// Enable the transit secrets engine
	path := "transit"
	sys := client.Sys()
	err = sys.Mount(path, &api.MountInput{
		Type:        "transit",
		Description: "Transit secrets engine for testing",
	})
	require.Nil(t, err, "Failed to mount transit secrets engine")

	return client, path
}

func TestHashicorpKMS(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, path := setupTestVault(t)

	// Use mock config instead of loading from environment
	mockConfig := &config.Config{
		VaultAddr:  "http://localhost:8200",
		VaultToken: "hvs.MlbuTDJ6DqdtgnFNuiOmNKej",
		VaultPath:  "transit",
	}

	// Initialize KMS service using test-only constructor
	NewHashiCorpKMS(mockConfig)
	kms := NewHashiCorpKMSWithClient(client, path)

	require.NotNil(t, kms, "Failed to initialize HashiCorpKMS")

	// Test StoreKEK
	t.Run("StoreKEK", func(t *testing.T) {
		kekID := "test-key-" + time.Now().Format("20060102150405")
		err := kms.StoreKEK(kekID, []byte("test-key"))
		require.Nil(t, err, "Failed to store KEK")
	})

	// Test RetrieveKEK
	t.Run("RetrieveKEK", func(t *testing.T) {
		kekID := "test-key-" + time.Now().Format("20060102150405")

		// First store a key
		err := kms.StoreKEK(kekID, []byte("test-key"))
		require.Nil(t, err, "Failed to store KEK")

		// Then retrieve it
		kek, err := kms.RetrieveKEK(kekID)
		require.Nil(t, err, "Failed to retrieve KEK")
		assert.NotEmpty(t, kek, "Retrieved KEK should not be empty")
	})

	// Cleanup
	t.Cleanup(func() {
		err := client.Sys().Unmount(path)
		if err != nil {
			t.Logf("Failed to unmount transit secrets engine: %v", err)
		}
	})
}
