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

func setupTestVault(t *testing.T) (*api.Client, string) {
	// Create a Vault client
	config := api.DefaultConfig()
	config.Address = "http://localhost:8200" // Default Vault dev server address

	client, err := api.NewClient(config)
	require.NoError(t, err, "Failed to create Vault client")

	// Set the root token (for dev server)
	client.SetToken("hvs.hHZbjEiA7AIbVBToXd5gc6gM")

	// Enable the transit secrets engine
	path := "transit"
	sys := client.Sys()
	err = sys.Mount(path, &api.MountInput{
		Type:        "transit",
		Description: "Transit secrets engine for testing",
	})
	require.NoError(t, err, "Failed to mount transit secrets engine")

	return client, path
}

func TestHashicorpKMS(t *testing.T) {

	cfg, _ := config.LoadConfig()
	// Initialize KMS service
	kms := NewHashiCorpKMS(cfg)
	// Test GenerateKEK
	t.Run("GenerateKEK", func(t *testing.T) {
		fmt.Println("\n--- Testing GenerateKEK ---")
		kek, err := kms.GenerateKEK()
		fmt.Println("\n--- Testing GenerateKEK ---", kek)
		require.NoError(t, err, "Failed to generate KEK")
		assert.NotEmpty(t, kek, "Generated KEK should not be empty")
		logger.LogEvent("test", "KEK_GENERATED", "test", "success", "Generated default KEK")
	})

	// Test StoreKEK
	t.Run("StoreKEK", func(t *testing.T) {
		fmt.Println("\n--- Testing StoreKEK ---")
		kekID := "test-key-" + time.Now().Format("20060102150405")
		fmt.Printf("Creating key with ID: %s\n", kekID)
		err := kms.StoreKEK(kekID, []byte("test-key"))
		require.NoError(t, err, "Failed to store KEK")
		logger.LogEvent("test", "KEK_STORED", "test", "success", fmt.Sprintf("Stored KEK with ID: %s", kekID))
	})

	// Test RetrieveKEK
	t.Run("RetrieveKEK", func(t *testing.T) {
		fmt.Println("\n--- Testing RetrieveKEK ---")
		kekID := "test-key-" + time.Now().Format("20060102150405")
		fmt.Printf("Creating and retrieving key with ID: %s\n", kekID)

		// First store a key
		err := kms.StoreKEK(kekID, []byte("test-key"))
		require.NoError(t, err, "Failed to store KEK")
		logger.LogEvent("test", "KEK_STORED", "test", "success", fmt.Sprintf("Stored KEK with ID: %s", kekID))

		// Then retrieve it
		kek, err := kms.RetrieveKEK(kekID)
		require.NoError(t, err, "Failed to retrieve KEK")
		assert.NotEmpty(t, kek, "Retrieved KEK should not be empty")
		logger.LogEvent("test", "KEK_RETRIEVED", "test", "success", fmt.Sprintf("Retrieved KEK with ID: %s", kekID))
	})

	// Test DeleteKEK
	t.Run("DeleteKEK", func(t *testing.T) {
		fmt.Println("\n--- Testing DeleteKEK ---")
		kekID := "test-key-" + time.Now().Format("20060102150405")
		fmt.Printf("Creating key for deletion test with ID: %s\n", kekID)

		// First store a key
		err := kms.StoreKEK(kekID, []byte("test-key"))
		require.NoError(t, err, "Failed to store KEK")
		logger.LogEvent("test", "KEK_STORED", "test", "success", fmt.Sprintf("Stored KEK with ID: %s", kekID))

		// Try to delete the key
		err = kms.DeleteKEK(kekID)
		if err != nil {
			// If deletion is not allowed, that's expected
			assert.Contains(t, err.Error(), "deletion is not allowed", "Expected error about deletion not being allowed")
			logger.LogEvent("test", "KEK_DE¸LETE_ATTEMPTED", "test", "expected_failure",
				fmt.Sprintf("Attempted to delete KEK with ID: %s (expected to fail as deletion is not allowed)", kekID))
		}

		// Verify the key still exists
		kek, err := kms.RetrieveKEK(kekID)
		require.NoError(t, err, "Key should still exist")
		assert.NotEmpty(t, kek, "Key should still be retrievable")
		logger.LogEvent("test", "KEK_VERIFIED", "test", "success", fmt.Sprintf("Verified KEK still exists with ID: %s", kekID))
	})

	// Test ListKEKs
	t.Run("ListKEKs", func(t *testing.T) {
		fmt.Println("\n--- Testing ListKEKs ---")
		// Create a few test keys
		var createdKeys []string
		for i := 0; i < 3; i++ {
			kekID := "test-key-" + time.Now().Format("20060102150405") + "-" + string(rune('a'+i))
			createdKeys = append(createdKeys, kekID)
			fmt.Printf("Creating key %d with ID: %s\n", i+1, kekID)
			err := kms.StoreKEK(kekID, []byte("test-key"))
			require.NoError(t, err, "Failed to store KEK")
			logger.LogEvent("test", "KEK_STORED", "test", "success", fmt.Sprintf("Stored KEK with ID: %s", kekID))
		}

		// List the keys
		keys, err := kms.ListKEKs()
		require.NoError(t, err, "Failed to list KEKs")
		assert.GreaterOrEqual(t, len(keys), 3, "Should find at least 3 keys")

		fmt.Println("\nListed keys:")
		for _, key := range keys {
			fmt.Printf("- %s\n", key)
		}
		logger.LogEvent("test", "KEYS_LISTED", "test", "success", fmt.Sprintf("Listed %d keys", len(keys)))
	})

	fmt.Println("\n=== Completed HashicorpKMS Tests ===")

	// Cleanup
	t.Cleanup(func() {
		// // Unmount the transit secrets engine
		// err := kms.Sys().Unmount(cfg.VaultPath)
		// if err != nil {
		// 	t.Logf("Failed to unmount transit secrets engine: %v", err)
		// }
		// logger.LogEvent("test", "CLEANUP_COMPLETE", "test", "success", "Unmounted transit secrets engine")
	})
}
