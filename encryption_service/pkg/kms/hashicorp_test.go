package kms

import (
	"fmt"
	"testing"
	"time"

	"encryption_microservice/internal/config"
	"encryption_microservice/pkg/logger"

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
	// Initialize logger
	logger.InitLogger()

	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	fmt.Println("\n=== Starting HashicorpKMS Tests ===")
	client, path := setupTestVault(t)

	config, _ := config.LoadConfig()
	// Initialize KMS service using test-only constructor
	NewHashiCorpKMS(config)
	kms := NewHashiCorpKMSWithClient(client, path)

	require.NotNil(t, kms, "Failed to initialize HashiCorpKMS")

	// Test StoreKEK
	t.Run("StoreKEK", func(t *testing.T) {
		fmt.Println("\n--- Testing StoreKEK ---")
		kekID := "test-key-" + time.Now().Format("20060102150405")
		fmt.Printf("Creating key with ID: %s\n", kekID)
		err := kms.StoreKEK(kekID, []byte("test-key"))
		require.Nil(t, err, "Failed to store KEK")
		logger.LogEvent("test", "KEK_STORED", "test", "success", fmt.Sprintf("Stored KEK with ID: %s", kekID))
	})

	// Test RetrieveKEK
	t.Run("RetrieveKEK", func(t *testing.T) {
		fmt.Println("\n--- Testing RetrieveKEK ---")
		kekID := "test-key-" + time.Now().Format("20060102150405")
		fmt.Printf("Creating and retrieving key with ID: %s\n", kekID)

		// First store a key
		err := kms.StoreKEK(kekID, []byte("test-key"))
		require.Nil(t, err, "Failed to store KEK")
		logger.LogEvent("test", "KEK_STORED", "test", "success", fmt.Sprintf("Stored KEK with ID: %s", kekID))

		// Then retrieve it
		kek, err := kms.RetrieveKEK(kekID)
		require.Nil(t, err, "Failed to retrieve KEK")
		assert.NotEmpty(t, kek, "Retrieved KEK should not be empty")
		logger.LogEvent("test", "KEK_RETRIEVED", "test", "success", fmt.Sprintf("Retrieved KEK with ID: %s", kekID))
	})

	// Test ListKEKs

	fmt.Println("\n=== Completed HashicorpKMS Tests ===")

	// Cleanup
	t.Cleanup(func() {
		err := client.Sys().Unmount(path)
		if err != nil {
			t.Logf("Failed to unmount transit secrets engine: %v", err)
		}
		logger.LogEvent("test", "CLEANUP_COMPLETE", "test", "success", "Unmounted transit secrets engine")
	})
}
