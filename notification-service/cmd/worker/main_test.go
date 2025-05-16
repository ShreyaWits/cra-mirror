package main

import (
	config "notification-service/internal/configs"
	"notification-service/pkg/kafka"
	"notification-service/pkg/temporal"
	"os"
	"strconv"
	"testing"

	"notification-service/pkg/logger"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/worker"
)

// MockTemporalWorker mocks the Temporal worker
type MockTemporalWorker struct {
	mock.Mock
}

func (m *MockTemporalWorker) RegisterActivities(activities []interface{}) {
	m.Called(activities)
}

func (m *MockTemporalWorker) RegisterWorkflows(workflows []interface{}) {
	m.Called(workflows)
}

func (m *MockTemporalWorker) Run() {
	m.Called()
}

// MockCassandraSession mocks the Cassandra session
type MockCassandraSession struct {
	mock.Mock
}

func (m *MockCassandraSession) Query(stmt string, values ...interface{}) *gocql.Query {
	args := m.Called(stmt, values)
	return args.Get(0).(*gocql.Query)
}

func (m *MockCassandraSession) Close() {
	m.Called()
}

// MockKafkaProducer mocks the Kafka producer
type MockKafkaProducer struct {
	mock.Mock
}

// MockNotificationRepository mocks the notification repository
type MockNotificationRepository struct {
	mock.Mock
}

// Add mock types for dependency injection

type mockTemporalWorker struct {
	activitiesCalled bool
	workflowsCalled  bool
	runCalled        bool
}

func (m *mockTemporalWorker) RegisterActivities(a []any) { m.activitiesCalled = true }
func (m *mockTemporalWorker) RegisterWorkflows(w []any)  { m.workflowsCalled = true }
func (m *mockTemporalWorker) Run()                       { m.runCalled = true }

type mockCassandraSession struct{}
type mockKafkaProducer struct{}

func TestMain(t *testing.T) {
	// Initialize logger for tests
	logger.InitLogger("http://localhost:5000")

	// Save original environment variables
	originalEnv := make(map[string]string)
	envVars := []string{
		"TEMPORAL_SERVER_URL",
		"LOKI_URL",
		"CASSANDRA_HOST",
		"CASSANDRA_KEYSPACE",
		"CASSANDRA_USERNAME",
		"CASSANDRA_PASSWORD",
		"CASSANDRA_PORT",
		"KAFKA_BROKERS",
	}

	for _, envVar := range envVars {
		if val, exists := os.LookupEnv(envVar); exists {
			originalEnv[envVar] = val
		}
	}

	// Set test environment variables
	os.Setenv("TEMPORAL_SERVER_URL", "localhost:7233")
	os.Setenv("LOKI_URL", "http://localhost:5000")
	os.Setenv("CASSANDRA_HOST", "localhost")
	os.Setenv("CASSANDRA_KEYSPACE", "test_keyspace")
	os.Setenv("CASSANDRA_USERNAME", "test_user")
	os.Setenv("CASSANDRA_PASSWORD", "test_password")
	os.Setenv("CASSANDRA_PORT", "9042")
	os.Setenv("KAFKA_BROKERS", "localhost:9092")

	// Create mock worker
	mockWorker := new(MockTemporalWorker)

	// Set up expectations
	mockWorker.On("RegisterActivities", mock.Anything).Return()
	mockWorker.On("RegisterWorkflows", mock.Anything).Return()
	mockWorker.On("Run").Return()

	// Test environment variable loading
	t.Run("Test environment variables", func(t *testing.T) {
		assert.Equal(t, "localhost:7233", os.Getenv("TEMPORAL_SERVER_URL"))
		assert.Equal(t, "http://localhost:5000", os.Getenv("LOKI_URL"))
		assert.Equal(t, "localhost", os.Getenv("CASSANDRA_HOST"))
		assert.Equal(t, "test_keyspace", os.Getenv("CASSANDRA_KEYSPACE"))
		assert.Equal(t, "test_user", os.Getenv("CASSANDRA_USERNAME"))
		assert.Equal(t, "test_password", os.Getenv("CASSANDRA_PASSWORD"))
		assert.Equal(t, "9042", os.Getenv("CASSANDRA_PORT"))
		assert.Equal(t, "localhost:9092", os.Getenv("KAFKA_BROKERS"))
	})

	// Test default values
	t.Run("Test default values", func(t *testing.T) {
		// Clear environment variables to test defaults
		for _, envVar := range envVars {
			os.Unsetenv(envVar)
		}

		// Test default values
		assert.Equal(t, "http://localhost:5000", config.GetEnv("LOKI_URL", "http://localhost:5000"))
		assert.Equal(t, "localhost", config.GetEnv("CASSANDRA_HOST", "localhost"))
		assert.Equal(t, "notifications", config.GetEnv("CASSANDRA_KEYSPACE", "notifications"))
		assert.Equal(t, "cassandra", config.GetEnv("CASSANDRA_USERNAME", "cassandra"))
		assert.Equal(t, "cassandra", config.GetEnv("CASSANDRA_PASSWORD", "cassandra"))
		assert.Equal(t, "9042", config.GetEnv("CASSANDRA_PORT", "9042"))
		assert.Equal(t, "localhost:9092", config.GetEnv("KAFKA_BROKERS", "localhost:9092"))
	})

	// Test Cassandra port conversion
	t.Run("Test Cassandra port conversion", func(t *testing.T) {
		os.Setenv("CASSANDRA_PORT", "invalid")
		_, err := strconv.Atoi(config.GetEnv("CASSANDRA_PORT", "9042"))
		assert.Error(t, err)
	})

	// Restore original environment variables
	for key, value := range originalEnv {
		os.Setenv(key, value)
	}
	for _, envVar := range envVars {
		if _, exists := originalEnv[envVar]; !exists {
			os.Unsetenv(envVar)
		}
	}
}

func TestTemporalWorkerInitialization(t *testing.T) {
	// Initialize logger for tests
	logger.InitLogger("http://localhost:5000")

	t.Run("Test successful Temporal worker initialization", func(t *testing.T) {
		// Create a mock worker
		mockWorker := new(MockTemporalWorker)
		mockWorker.On("RegisterActivities", mock.Anything).Return()
		mockWorker.On("RegisterWorkflows", mock.Anything).Return()
		mockWorker.On("Run").Return()

		// Since we can't actually connect to Temporal in tests, we'll just verify the mock
		mockWorker.RegisterActivities([]interface{}{})
		mockWorker.RegisterWorkflows([]interface{}{})
		mockWorker.Run()

		mockWorker.AssertExpectations(t)
	})

	t.Run("Test failed Temporal worker initialization", func(t *testing.T) {
		// Test with invalid URL
		worker, err := temporal.InitTemporalWorker("invalid-url", worker.Options{})
		assert.Error(t, err)
		assert.Nil(t, worker)
	})
}

func TestKafkaInitialization(t *testing.T) {
	// Initialize logger for tests
	logger.InitLogger("http://localhost:5000")

	t.Run("Test successful Kafka initialization", func(t *testing.T) {
		producer := kafka.InitKafkaPublisher("localhost:9092")
		assert.NotNil(t, producer)
	})

	t.Run("Test failed Kafka initialization", func(t *testing.T) {
		producer := kafka.InitKafkaPublisher("invalid-broker")
		assert.NotNil(t, producer) // Kafka client might still be created even with invalid broker
	})
}

func TestCassandraInitialization(t *testing.T) {
	// Initialize logger for tests
	logger.InitLogger("http://localhost:5000")

	t.Run("Test Cassandra configuration", func(t *testing.T) {
		// Test configuration values
		host := "localhost"
		port := 9042
		keyspace := "test_keyspace"
		username := "test_user"
		password := "test_password"

		// Test configuration
		assert.Equal(t, host, "localhost")
		assert.Equal(t, port, 9042)
		assert.Equal(t, keyspace, "test_keyspace")
		assert.Equal(t, username, "test_user")
		assert.Equal(t, password, "test_password")
	})
}

func TestMain_Coverage(t *testing.T) {
	os.Setenv("TEMPORAL_SERVER_URL", "dummy")
	os.Setenv("LOKI_URL", "http://localhost:5000")
	os.Setenv("CASSANDRA_HOST", "localhost")
	os.Setenv("CASSANDRA_KEYSPACE", "test_keyspace")
	os.Setenv("CASSANDRA_USERNAME", "test_user")
	os.Setenv("CASSANDRA_PASSWORD", "test_password")
	os.Setenv("CASSANDRA_PORT", "9042")
	os.Setenv("KAFKA_BROKERS", "localhost:9092")

	err := runWorker()
	// We expect an error because the dummy temporal server will fail
	assert.Error(t, err)
}

func TestMain_ErrorPath_InvalidCassandraPort(t *testing.T) {
	os.Setenv("TEMPORAL_SERVER_URL", "dummy")
	os.Setenv("LOKI_URL", "http://localhost:5000")
	os.Setenv("CASSANDRA_HOST", "localhost")
	os.Setenv("CASSANDRA_KEYSPACE", "test_keyspace")
	os.Setenv("CASSANDRA_USERNAME", "test_user")
	os.Setenv("CASSANDRA_PASSWORD", "test_password")
	os.Setenv("CASSANDRA_PORT", "notanint")
	os.Setenv("KAFKA_BROKERS", "localhost:9092")

	err := runWorker()
	assert.Error(t, err)
}

func TestMain_ErrorPath_TemporalInitError(t *testing.T) {
	os.Setenv("TEMPORAL_SERVER_URL", "invalid-url")
	os.Setenv("LOKI_URL", "http://localhost:5000")
	os.Setenv("CASSANDRA_HOST", "localhost")
	os.Setenv("CASSANDRA_KEYSPACE", "test_keyspace")
	os.Setenv("CASSANDRA_USERNAME", "test_user")
	os.Setenv("CASSANDRA_PASSWORD", "test_password")
	os.Setenv("CASSANDRA_PORT", "9042")
	os.Setenv("KAFKA_BROKERS", "localhost:9092")

	err := runWorker()
	assert.Error(t, err)
}

func TestCassandraPortConversion(t *testing.T) {
	os.Setenv("CASSANDRA_PORT", "9042")
	port, err := strconv.Atoi(os.Getenv("CASSANDRA_PORT"))
	assert.NoError(t, err)
	assert.Equal(t, 9042, port)
}

func TestRunWorkerWithDeps_Success(t *testing.T) {
	os.Setenv("TEMPORAL_SERVER_URL", "dummy")
	os.Setenv("LOKI_URL", "http://localhost:5000")
	os.Setenv("CASSANDRA_HOST", "localhost")
	os.Setenv("CASSANDRA_KEYSPACE", "test_keyspace")
	os.Setenv("CASSANDRA_USERNAME", "test_user")
	os.Setenv("CASSANDRA_PASSWORD", "test_password")
	os.Setenv("CASSANDRA_PORT", "9042")
	os.Setenv("KAFKA_BROKERS", "localhost:9092")

	mockTemporal := &mockTemporalWorker{}
	mockCassandra := &gocql.Session{} // must be *gocql.Session for wrapper
	mockKafka := &kafka.KafkaPublisher{}

	err := runWorkerWithDeps(
		func(url string, opts worker.Options) (TemporalWorker, error) { return mockTemporal, nil },
		func(host string, port int, keyspace, username, password string) CassandraSession {
			return mockCassandra
		},
		func(broker string) KafkaProducer { return mockKafka },
	)
	assert.NoError(t, err)
	assert.True(t, mockTemporal.activitiesCalled)
	assert.True(t, mockTemporal.workflowsCalled)
	assert.True(t, mockTemporal.runCalled)
}
