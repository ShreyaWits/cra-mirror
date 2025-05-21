package cassandra

import (
	"notification-service/pkg/config"
	"strconv"
	"testing"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
)

// Mock MigrationRunner
type mockMigrationRunner struct {
	called bool
}

func (m *mockMigrationRunner) RunMigrations(session *gocql.Session) {
	m.called = true
}

func mockCreateSession(cluster *gocql.ClusterConfig) (*gocql.Session, error) {
	return &gocql.Session{}, nil
}

func TestNewCassandraDB(t *testing.T) {
	CASSANDRA_HOST := config.GetEnv("CASSANDRA_HOST", "localhost")
	CASSANDRA_KEYSPACE := config.GetEnv("CASSANDRA_KEYSPACE", "notifications")
	CASSANDRA_USERNAME := config.GetEnv("CASSANDRA_USERNAME", "cassandra")
	CASSANDRA_PASSWORD := config.GetEnv("CASSANDRA_PASSWORD", "cassandra")
	CASSANDRA_PORT, err := strconv.Atoi(config.GetEnv("CASSANDRA_PORT", "9042"))

	if err != nil {
		panic(err)
	}

	mockRunner := &mockMigrationRunner{}
	session := newCassandraDB(
		CASSANDRA_HOST, CASSANDRA_PORT, CASSANDRA_KEYSPACE, CASSANDRA_USERNAME, CASSANDRA_PASSWORD,
		func(h string) *gocql.ClusterConfig { return &gocql.ClusterConfig{} },
		mockRunner.RunMigrations,
	)

	assert.NotNil(t, session)
	assert.True(t, mockRunner.called, "Expected RunMigrations to be called")
}

func TestNewCassandraDB_SessionError(t *testing.T) {
	mockRunner := &mockMigrationRunner{}
	mockNewCluster := func(host string) *gocql.ClusterConfig {
		return &gocql.ClusterConfig{}
	}

	calledFatal := false
	mockFatalf := func(format string, v ...interface{}) {
		calledFatal = true
		panic("fatal")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic from log.Fatalf, got none")
		}
		if !calledFatal {
			t.Errorf("Expected logFatalf to be called")
		}
	}()

	newCassandraDBWithFatalf("localhost", 9042, "ks", "user", "pass", mockNewCluster, mockRunner.RunMigrations, mockFatalf)
}

func TestNewCassandraDB_DependencyInjection(t *testing.T) {
	mockRunner := &mockMigrationRunner{}
	mockNewCluster := func(host string) *gocql.ClusterConfig {
		return &gocql.ClusterConfig{Keyspace: "ks"}
	}

	session := newCassandraDB("localhost", 9042, "ks", "user", "pass", mockNewCluster, mockRunner.RunMigrations)
	assert.NotNil(t, session)
	assert.True(t, mockRunner.called, "Expected RunMigrations to be called")
}
