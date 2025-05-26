package repositories

import (
	"context"
	"fmt"
	etcdDB "nps-config-service/pkg/etcd"
	"nps-config-service/pkg/observability"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestConfigRepository tests the ConfigRepository interacting with a real Etcd container
func TestConfigRepository(t *testing.T) {
	ctx := context.Background()

	// Initialize observability stack for testing
	observabilityStack := observability.NewObservabilityStack("config-service-test")

	req := testcontainers.ContainerRequest{
		Image:        "gcr.io/etcd-development/etcd:v3.5.14",
		ExposedPorts: []string{"2379/tcp"},
		WaitingFor:   wait.ForListeningPort("2379/tcp").WithStartupTimeout(10 * time.Second),
		Cmd: []string{
			"/usr/local/bin/etcd",
			"--advertise-client-urls", "http://0.0.0.0:2379",
			"--listen-client-urls", "http://0.0.0.0:2379",
		},
	}

	etcdContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer func() {
		_ = etcdContainer.Terminate(ctx)
	}()

	host, err := etcdContainer.Host(ctx)
	require.NoError(t, err)

	port, err := etcdContainer.MappedPort(ctx, "2379")
	require.NoError(t, err)

	endpoint := fmt.Sprintf("%s:%s", host, port.Port())
	fmt.Println("Etcd endpoint:", endpoint)
	time.Sleep(2 * time.Second)

	// Initialize Etcd client
	client, err := etcdDB.InitEtcdDB(endpoint)
	require.NoError(t, err)

	imp := etcdDB.NewEtcdClientImpl(client)
	require.NotNil(t, imp, "Failed to create Etcd client")

	// Create ConfigRepository with Etcd client and observability stack
	repo := NewConfigRepository(imp, observabilityStack)

	// Test data
	serviceName := "configservice"
	environment := "prod"
	configData := map[string]interface{}{
		"db_url":      "postgres://user:pass@localhost:5432/db",
		"retry_count": "3",
		"timeout":     "5000",
	}

	// Test StoreConfig
	t.Run("StoreConfig", func(t *testing.T) {
		_, err := repo.StoreConfig(ctx, serviceName, environment, configData)
		require.NoError(t, err)
	})

	// Test GetConfig
	t.Run("GetConfig", func(t *testing.T) {
		retrievedConfig, err := repo.GetConfig(ctx, serviceName, environment)
		require.NoError(t, err)
		assert.Equal(t, configData, retrievedConfig)
	})

	// Test GetConfigValue
	t.Run("GetConfigValue", func(t *testing.T) {
		value, err := repo.GetConfigValue(ctx, serviceName, environment, "db_url")
		require.NoError(t, err)
		assert.Equal(t, "postgres://user:pass@localhost:5432/db", value)
	})

	// Test SetEtcdKey
	t.Run("SetEtcdKey", func(t *testing.T) {
		key := "test/key"
		value := "test-value"
		err := repo.SetEtcdKey(ctx, key, value, 0) // 0 TTL means no expiration
		require.NoError(t, err)

		// Verify the key was set
		retrievedValue, err := repo.GetEtcdKey(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, retrievedValue)
	})

	// Test DeleteEtcdKey
	t.Run("DeleteEtcdKey", func(t *testing.T) {
		key := "test/key/to/delete"
		value := "delete-me"

		// First set a key
		err := repo.SetEtcdKey(ctx, key, value, 0) // 0 TTL means no expiration
		require.NoError(t, err)

		// Then delete it
		err = repo.DeleteEtcdKey(ctx, key)
		require.NoError(t, err)

		// Verify it's deleted
		_, err = repo.GetEtcdKey(ctx, key)
		assert.Error(t, err, "Expected error when getting deleted key")
	})
}
