package app

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"notification-service/internal/common/api/dtos"
	"notification-service/internal/common/models"
	"notification-service/internal/common/repositories"
	"notification-service/pkg/di"
	"notification-service/pkg/kafka"
	"notification-service/pkg/redis"
	"notification-service/pkg/temporal"
)

// --- Mocks ---
type MockSession struct{ mock.Mock }

func (m *MockSession) Query(stmt string, values ...interface{}) repositories.QueryExecutor {
	args := m.Called(stmt, values)
	return args.Get(0).(repositories.QueryExecutor)
}
func (m *MockSession) ExecuteBatch(batch *gocql.Batch) error {
	args := m.Called(batch)
	return args.Error(0)
}

type MockQuery struct{ mock.Mock }

func (m *MockQuery) Exec() error                    { return m.Called().Error(0) }
func (m *MockQuery) Iter() *gocql.Iter              { return m.Called().Get(0).(*gocql.Iter) }
func (m *MockQuery) Scan(dest ...interface{}) error { return m.Called(dest...).Error(0) }

type MockConfigRepo struct{}

// Implement required interfaces for mocks
func (m *MockConfigRepo) SaveConfig(ctx context.Context, configs []dtos.ChannelConfig) error {
	return nil
}
func (m *MockConfigRepo) GetConfig() ([]dtos.ChannelConfig, error) { return nil, nil }
func (m *MockConfigRepo) GetConfigByChannel(channel string) (*dtos.ChannelConfig, error) {
	return nil, nil
}

type MockNotificationRepo struct{}

func (m *MockNotificationRepo) SaveNotification(notification *models.Notification) (*string, error) {
	return nil, nil
}
func (m *MockNotificationRepo) UpdateNotificationByID(notificationID string, updateData *models.Notification) error {
	return nil
}

type MockKafkaPublisher struct{}
type MockTemporalClient struct{}

type MockContainer struct {
	RegisterCalls []interface{}
}

func (m *MockContainer) RegisterService(service interface{}) {
	m.RegisterCalls = append(m.RegisterCalls, service)
}

// Add a mock for RedisRepositoryInterface
type MockRedisRepo struct{}

func (m *MockRedisRepo) GetNotificationConfig(key string) ([]dtos.ChannelConfig, error) {
	return nil, nil
}
func (m *MockRedisRepo) SetNotificationConfig(key string, config []dtos.ChannelConfig) error {
	return nil
}

// --- Test helpers ---
func setEnvVars(vars map[string]string) func() {
	old := map[string]string{}
	for k, v := range vars {
		old[k] = os.Getenv(k)
		os.Setenv(k, v)
	}
	return func() {
		for k, v := range old {
			os.Setenv(k, v)
		}
	}
}

// --- Tests ---
// func TestCassandraSessionWrapper_Query(t *testing.T) {
// 	mockSession := new(MockSession)
// 	mockQuery := new(MockQuery)
// 	mockSession.On("Query", "SELECT 1", []interface{}{}).Return(mockQuery)
// 	w := &cassandraSessionWrapper{Session: mockSession}
// 	q := w.Query("SELECT 1")
// 	assert.NotNil(t, q)
// 	mockSession.AssertExpectations(t)
// }

// func TestQueryExecutorWrapper_Exec(t *testing.T) {
// 	mockQuery := new(MockQuery)
// 	mockQuery.On("Exec").Return(nil)
// 	w := &queryExecutorWrapper{Query: mockQuery}
// 	err := w.Exec()
// 	assert.NoError(t, err)
// 	mockQuery.AssertExpectations(t)
// }

func TestInitDependency_Success(t *testing.T) {
	reset := setEnvVars(map[string]string{
		"CASSANDRA_HOST":      "localhost",
		"CASSANDRA_KEYSPACE":  "ks",
		"CASSANDRA_USERNAME":  "user",
		"CASSANDRA_PASSWORD":  "pass",
		"CASSANDRA_PORT":      "9042",
		"REDIS_HOST":          "localhost",
		"REDIS_PORT":          "6379",
		"TEMPORAL_SERVER_URL": "localhost:7233",
		"TEMPORAL_QUEUE":      "queue",
		"KAFKA_BROKERS":       "localhost:9092",
	})
	defer reset()

	container = nil // reset global
	origInit := initContainer
	initContainer = func() *di.Container {
		return di.InitContainer()
	}
	defer func() { initContainer = origInit }()

	// Patch dependency constructors
	origNewCassandraDB := newCassandraDB
	newCassandraDB = func(host string, port int, keyspace, username, password string) *gocql.Session {
		return &gocql.Session{}
	}
	defer func() { newCassandraDB = origNewCassandraDB }()

	origInitRedisService := initRedisService
	initRedisService = func(host, port, user, pass string) (*redis.RedisService, error) {
		return &redis.RedisService{}, nil
	}
	defer func() { initRedisService = origInitRedisService }()

	origNewRedisService := newRedisService
	newRedisService = func(redisService redis.RedisServiceInterface) *repositories.NotificationRedis {
		return repositories.InitRedisRepo(redisService)
	}
	defer func() { newRedisService = origNewRedisService }()

	origNewConfigRepo := newConfigRepository
	newConfigRepository = func(session repositories.CassandraSession) (repositories.ConfigRepositoryInterface, error) {
		return &MockConfigRepo{}, nil
	}
	defer func() { newConfigRepository = origNewConfigRepo }()

	origNewNotificationRepo := newNotificationRepository
	newNotificationRepository = func(session repositories.CassandraSession) repositories.NotificationRepositoryInterface {
		return &MockNotificationRepo{}
	}
	defer func() { newNotificationRepository = origNewNotificationRepo }()

	origInitKafkaPublisher := initKafkaPublisher
	initKafkaPublisher = func(brokers string) *kafka.KafkaPublisher {
		return &kafka.KafkaPublisher{}
	}
	defer func() { initKafkaPublisher = origInitKafkaPublisher }()

	origInitTemporal := initTemporal
	initTemporal = func(url, queue, svc string, repo repositories.NotificationRepositoryInterface, kafka *kafka.KafkaPublisher) (*temporal.TemporalClient, error) {
		return &temporal.TemporalClient{}, nil
	}
	defer func() { initTemporal = origInitTemporal }()

	InitDependency()
	assert.NotNil(t, container)
}

func TestInitDependency_CassandraPortError(t *testing.T) {
	reset := setEnvVars(map[string]string{"CASSANDRA_PORT": "notanint"})
	defer reset()
	container = nil
	origInit := initContainer
	initContainer = func() *di.Container {
		return di.InitContainer()
	}
	defer func() { initContainer = origInit }()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic on invalid port")
		}
	}()
	InitDependency()
}

func TestInitDependency_DependencyErrors(t *testing.T) {
	reset := setEnvVars(map[string]string{"CASSANDRA_PORT": "9042"})
	defer reset()
	container = nil
	origInit := initContainer
	initContainer = func() *di.Container {
		return di.InitContainer()
	}
	defer func() { initContainer = origInit }()

	// Patch dependency constructors to return errors
	origNewCassandraDB := newCassandraDB
	newCassandraDB = func(host string, port int, keyspace, username, password string) *gocql.Session {
		return &gocql.Session{}
	}
	defer func() { newCassandraDB = origNewCassandraDB }()

	origInitRedisService := initRedisService
	initRedisService = func(host, port, user, pass string) (*redis.RedisService, error) {
		return nil, errors.New("fail")
	}
	defer func() { initRedisService = origInitRedisService }()

	origNewRedisService := newRedisService
	newRedisService = func(redisService redis.RedisServiceInterface) *repositories.NotificationRedis {
		return repositories.InitRedisRepo(redisService)
	}
	defer func() { newRedisService = origNewRedisService }()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic on redis error")
		}
	}()
	InitDependency()
}

// For DiService getter tests, create a minimal mock container with InvokeService

type MockInvokeContainer struct {
	InvokeFunc func(interface{})
}

func (m *MockInvokeContainer) InvokeService(service interface{}) {
	m.InvokeFunc(service)
}

func TestDiService_Getters(t *testing.T) {
	realContainer := di.InitContainer()
	realContainer.RegisterService(func() *gocql.Session { return &gocql.Session{} })
	realContainer.RegisterService(func() repositories.RedisRepositoryInterface {
		return repositories.InitRedisRepo(&redis.RedisService{})
	})
	realContainer.RegisterService(func() repositories.ConfigRepositoryInterface { return &MockConfigRepo{} })
	realContainer.RegisterService(func() repositories.NotificationRepositoryInterface { return &MockNotificationRepo{} })
	realContainer.RegisterService(func() *temporal.TemporalClient { return &temporal.TemporalClient{} })
	realContainer.RegisterService(func() *kafka.KafkaPublisher { return &kafka.KafkaPublisher{} })

	d := &DiService{Container: realContainer}
	assert.NotNil(t, d.GetCassandraSession())
	assert.NotNil(t, d.GetRedisService())
	assert.NotNil(t, d.GetConfigRepo())
	assert.NotNil(t, d.GetNotificationRepo())
	assert.NotNil(t, d.GetTemporalClient())
	assert.NotNil(t, d.GetKafka())
}

func TestInit(t *testing.T) {
	container = &di.Container{}
	d := Init()
	assert.NotNil(t, d)
	assert.Equal(t, container, d.Container)
}

func TestGetContainer(t *testing.T) {
	container = &di.Container{}
	assert.Equal(t, container, GetContainer())
}
