package main

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"notification-service/internal/app"
	"notification-service/internal/common/api/dtos"
	"notification-service/internal/common/models"
	"notification-service/internal/common/repositories"
	"notification-service/pkg/di"
	"notification-service/pkg/kafka"
	"notification-service/pkg/temporal"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type testLogger struct {
	fatalCalled   bool
	errorCalled   bool
	infoCalled    bool
	printlnCalled bool
}

func (l *testLogger) Println(args ...interface{})               { l.printlnCalled = true }
func (l *testLogger) Info(args ...interface{})                  { l.infoCalled = true }
func (l *testLogger) Errorf(format string, args ...interface{}) { l.errorCalled = true }
func (l *testLogger) Fatalf(format string, args ...interface{}) { l.fatalCalled = true }

type testGRPCServer struct {
	startErr error
	started  bool
	stopped  bool
	onStart  func()
	server   *grpc.Server
}

func (s *testGRPCServer) Start() error {
	s.started = true
	if s.onStart != nil {
		s.onStart()
	}
	return s.startErr
}
func (s *testGRPCServer) Stop() { s.stopped = true }
func (s *testGRPCServer) GetServer() *grpc.Server {
	if s.server == nil {
		s.server = grpc.NewServer()
	}
	return s.server
}

type testKafkaConsumer struct {
	closed   bool
	consumed bool
	started  bool
	onStart  func()
}

func (c *testKafkaConsumer) Close()                                    { c.closed = true }
func (c *testKafkaConsumer) Consume(topic string, handler interface{}) { c.consumed = true }
func (c *testKafkaConsumer) Start(ctx context.Context) {
	c.started = true
	if c.onStart != nil {
		c.onStart()
	}
}

// Mock for RedisRepositoryInterface
type MockRedisRepo struct{}

func (m *MockRedisRepo) GetNotificationConfig(key string) ([]dtos.ChannelConfig, error) {
	return nil, nil
}
func (m *MockRedisRepo) SetNotificationConfig(key string, config []dtos.ChannelConfig) error {
	return nil
}

// Mock for ConfigRepositoryInterface
type MockConfigRepo struct{}

func (m *MockConfigRepo) SaveConfig(ctx context.Context, configs []dtos.ChannelConfig) error {
	return nil
}
func (m *MockConfigRepo) GetConfig() ([]dtos.ChannelConfig, error) { return nil, nil }
func (m *MockConfigRepo) GetConfigByChannel(channel string) (*dtos.ChannelConfig, error) {
	return nil, nil
}

// Mock for NotificationRepositoryInterface
type MockNotificationRepo struct{}

func (m *MockNotificationRepo) SaveNotification(notification *models.Notification) (*string, error) {
	return nil, nil
}
func (m *MockNotificationRepo) UpdateNotificationByID(notificationID string, updateData *models.Notification) error {
	return nil
}

func TestRunServer_Success(t *testing.T) {
	container := di.InitContainer()
	container.RegisterService(func() repositories.RedisRepositoryInterface { return &MockRedisRepo{} })
	container.RegisterService(func() repositories.ConfigRepositoryInterface { return &MockConfigRepo{} })
	container.RegisterService(func() repositories.NotificationRepositoryInterface { return &MockNotificationRepo{} })
	container.RegisterService(func() *temporal.TemporalClient { return &temporal.TemporalClient{} })
	container.RegisterService(func() *kafka.KafkaPublisher { return &kafka.KafkaPublisher{} })

	// Patch the global container in the app package
	app.SetContainer(container)

	logger := &testLogger{}
	var wg sync.WaitGroup
	wg.Add(2)
	grpcServer := &testGRPCServer{onStart: wg.Done}
	kafkaConsumer := &testKafkaConsumer{onStart: wg.Done}
	shutdownCalled := false

	err := runServer(
		func(key, def string) string { return def },
		func() {},
		func(port string) GRPCServer { return grpcServer },
		func(brokers []string, groupID string, topics []string) (KafkaConsumer, error) {
			return kafkaConsumer, nil
		},
		logger,
		func(c chan<- os.Signal, sig ...os.Signal) { go func() { c <- os.Interrupt }() },
		func(d time.Duration) {},
	)
	wg.Wait()
	assert.NoError(t, err)
	assert.True(t, grpcServer.started)
	assert.True(t, kafkaConsumer.started)
	assert.True(t, logger.infoCalled)
	assert.True(t, shutdownCalled)
}

func TestRunServer_KafkaError(t *testing.T) {
	logger := &testLogger{}
	err := runServer(
		func(key, def string) string { return def },
		func() {},
		func(port string) GRPCServer { return &testGRPCServer{} },
		func(brokers []string, groupID string, topics []string) (KafkaConsumer, error) {
			return nil, errors.New("kafka error")
		},
		logger,
		func(c chan<- os.Signal, sig ...os.Signal) {},
		func(d time.Duration) {},
	)
	assert.Error(t, err)
	assert.True(t, logger.fatalCalled)
}

func TestRunServer_GRPCError(t *testing.T) {
	logger := &testLogger{}
	var wg sync.WaitGroup
	wg.Add(1)
	grpcServer := &testGRPCServer{startErr: errors.New("grpc error"), onStart: wg.Done}
	kafkaConsumer := &testKafkaConsumer{}

	err := runServer(
		func(key, def string) string { return def },
		func() {},
		func(port string) GRPCServer { return grpcServer },
		func(brokers []string, groupID string, topics []string) (KafkaConsumer, error) {
			return kafkaConsumer, nil
		},
		logger,
		func(c chan<- os.Signal, sig ...os.Signal) { go func() { c <- os.Interrupt }() },
		func(d time.Duration) {},
	)
	wg.Wait()
	assert.NoError(t, err) // gRPC error is logged, not returned
	assert.True(t, grpcServer.started)
	assert.True(t, grpcServer.stopped)
	assert.True(t, logger.errorCalled)
}

func TestRunServer_GracefulShutdown(t *testing.T) {
	logger := &testLogger{}
	var wg sync.WaitGroup
	wg.Add(2)
	grpcServer := &testGRPCServer{onStart: wg.Done}
	kafkaConsumer := &testKafkaConsumer{onStart: wg.Done}
	shutdownCalled := false

	shutdown := make(chan struct{})
	// Simulate signal after a short delay
	err := runServer(
		func(key, def string) string { return def },
		func() {},
		func(port string) GRPCServer { return grpcServer },
		func(brokers []string, groupID string, topics []string) (KafkaConsumer, error) {
			return kafkaConsumer, nil
		},
		logger,
		func(c chan<- os.Signal, sig ...os.Signal) {
			go func() { time.Sleep(10 * time.Millisecond); c <- os.Interrupt }()
		},
		func(d time.Duration) { close(shutdown) },
	)
	wg.Wait()
	<-shutdown
	assert.NoError(t, err)
	assert.True(t, logger.infoCalled)
	assert.True(t, shutdownCalled)
}

func TestRunServer_KafkaErrorWithShutdown(t *testing.T) {
	logger := &testLogger{}
	shutdownCalled := false
	err := runServer(
		func(key, def string) string { return def },
		func() {},
		func(port string) GRPCServer { return &testGRPCServer{} },
		func(brokers []string, groupID string, topics []string) (KafkaConsumer, error) {
			return nil, errors.New("kafka error")
		},
		logger,
		func(c chan<- os.Signal, sig ...os.Signal) {},
		func(d time.Duration) {},
	)
	assert.Error(t, err)
	assert.True(t, logger.fatalCalled)
	assert.True(t, shutdownCalled)
}
