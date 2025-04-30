package repositories

import (
	"context"
	"fmt"
	etcdDB "nps-config-service/pkg/etcd"
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
	// endpoint, err := etcdContainer.ClientEndpoint(ctx)
	client, err := etcdDB.InitEtcdDB(endpoint) // Removed or commented out as etcdDB is undefined
	// etcdDB.InitEtcdDB(etcdContainer.ClientEndpoint()) // Removed or commented out as etcdDB is undefined
	imp := etcdDB.NewEtcdClientImpl(client)
	// Step 3: Create ConfigRepository with Etcd client
	if imp == nil {
		t.Fatalf("Failed to create Etcd client")
	}
	repo := NewConfigRepository(imp)

	// Step 4: Store configuration data
	serviceName := "configservice"
	environment := "prod"
	configData := map[string]interface{}{
		"db_url":      "postgres://user:pass@localhost:5432/db",
		"retry_count": "3",
		"timeout":     "5000",
	}

	// Act: Store config in Etcd
	_, err = repo.StoreConfig(serviceName, environment, configData)
	require.NoError(t, err)

	// Step 5: Retrieve the stored configuration data
	retrievedConfig, err := repo.GetConfig(serviceName, environment)
	require.NoError(t, err)

	// Assert: Check that the stored and retrieved config match
	assert.Equal(t, configData, retrievedConfig)

	// Clean up: The Etcd container is automatically cleaned up by `defer`
}
