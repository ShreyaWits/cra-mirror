package mock

import (
	"context"
	"reflect"
	"testing"
	"time"

	"messaging_service/pkg/confluent"
	"messaging_service/pkg/observability"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
)

// SetupMockDLQProducer creates a DLQProducerImpl with mocked dependencies
// Uses testify/mock style mocks for compatibility with existing tests
func SetupMockDLQProducer(t *testing.T) (*confluent.DLQProducerImpl, *MockKafkaProducerTestify, *MockObservabilityStackTestify) {
	mockProducer := NewMockKafkaProducerTestify()
	mockObs := NewMockObservabilityStackTestify()

	// Setup common expectations
	mockObs.TracerService.On("StartTracer", mock.Anything, mock.Anything).Return(context.Background(), "span")
	mockObs.TracerService.On("StopSpan", mock.Anything).Return()

	// Setup default return for Events() to avoid nil pointer dereference
	eventsChan := make(chan kafka.Event, 1)
	mockProducer.On("Events").Return(eventsChan).Maybe()

	// Create a basic config for testing
	cfg := confluent.KafkaConfig{
		Topic:        "test-topic",
		WriteTimeout: 5 * time.Second,
	}

	// Create a DLQProducerImpl
	dlqProducer := &confluent.DLQProducerImpl{}

	// Use reflection to set private fields
	dlqProducerValue := reflect.ValueOf(dlqProducer).Elem()

	// Set Producer field
	producerField := dlqProducerValue.FieldByName("Producer")
	if producerField.IsValid() && producerField.CanSet() {
		producerField.Set(reflect.ValueOf(mockProducer))
	}

	// Set Topic field
	topicField := dlqProducerValue.FieldByName("Topic")
	if topicField.IsValid() && topicField.CanSet() {
		topicField.Set(reflect.ValueOf("test-topic-dlq"))
	}

	// Set Config field
	configField := dlqProducerValue.FieldByName("Config")
	if configField.IsValid() && configField.CanSet() {
		configField.Set(reflect.ValueOf(cfg))
	}

	// Set obs field
	obsField := dlqProducerValue.FieldByName("obs")
	if obsField.IsValid() && obsField.CanSet() {
		obsField.Set(reflect.ValueOf(mockObs))
	}

	return dlqProducer, mockProducer, mockObs
}

// SetupMockDLQProducerGomock creates a DLQProducerImpl with gomock mocked dependencies
// Uses golang/mock style mocks for compatibility with gomock-generated mocks
func SetupMockDLQProducerGomock(t *testing.T, ctrl *gomock.Controller) (*confluent.DLQProducerImpl, *MockKafkaProducerInterface, *observability.ObservabilityStack) {
	mockProducer := NewMockKafkaProducerInterface(ctrl)

	// Create a basic config for testing
	cfg := confluent.KafkaConfig{
		Topic:        "test-topic",
		WriteTimeout: 5 * time.Second,
	}

	// For gomock testing, we'll use a real observability stack
	// In a real test, you'd likely want to mock this further
	obs := &observability.ObservabilityStack{}

	// Create a DLQProducerImpl
	dlqProducer := &confluent.DLQProducerImpl{}

	// Use reflection to set private fields
	dlqProducerValue := reflect.ValueOf(dlqProducer).Elem()

	// Set Producer field
	producerField := dlqProducerValue.FieldByName("Producer")
	if producerField.IsValid() && producerField.CanSet() {
		producerField.Set(reflect.ValueOf(mockProducer))
	}

	// Set Topic field
	topicField := dlqProducerValue.FieldByName("Topic")
	if topicField.IsValid() && topicField.CanSet() {
		topicField.Set(reflect.ValueOf("test-topic-dlq"))
	}

	// Set Config field
	configField := dlqProducerValue.FieldByName("Config")
	if configField.IsValid() && configField.CanSet() {
		configField.Set(reflect.ValueOf(cfg))
	}

	// Set obs field
	obsField := dlqProducerValue.FieldByName("obs")
	if obsField.IsValid() && obsField.CanSet() {
		obsField.Set(reflect.ValueOf(obs))
	}

	return dlqProducer, mockProducer, obs
}

// MockKafkaProducerTestify is a testify/mock implementation of KafkaProducerInterface for tests
type MockKafkaProducerTestify struct {
	mock.Mock
}

func NewMockKafkaProducerTestify() *MockKafkaProducerTestify {
	return &MockKafkaProducerTestify{}
}

func (m *MockKafkaProducerTestify) Events() chan kafka.Event {
	args := m.Called()
	return args.Get(0).(chan kafka.Event)
}

func (m *MockKafkaProducerTestify) Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error {
	args := m.Called(msg, deliveryChan)
	return args.Error(0)
}

func (m *MockKafkaProducerTestify) Flush(timeoutMs int) int {
	args := m.Called(timeoutMs)
	return args.Int(0)
}

func (m *MockKafkaProducerTestify) InitTransactions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockKafkaProducerTestify) BeginTransaction() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockKafkaProducerTestify) CommitTransaction(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockKafkaProducerTestify) AbortTransaction(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockKafkaProducerTestify) Close() {
	m.Called()
}

// MockObservabilityStackTestify is a testify/mock implementation of ObservabilityStack for tests
type MockObservabilityStackTestify struct {
	TracerService *MockTracerServiceTestify
	LoggerService *MockLoggerServiceTestify
}

func NewMockObservabilityStackTestify() *MockObservabilityStackTestify {
	return &MockObservabilityStackTestify{
		TracerService: new(MockTracerServiceTestify),
		LoggerService: new(MockLoggerServiceTestify),
	}
}

type MockTracerServiceTestify struct {
	mock.Mock
}

func (m *MockTracerServiceTestify) StartTracer(ctx context.Context, functionName string) (context.Context, interface{}) {
	args := m.Called(ctx, functionName)
	return args.Get(0).(context.Context), args.Get(1)
}

func (m *MockTracerServiceTestify) StopSpan(span interface{}) {
	m.Called(span)
}

type MockLoggerServiceTestify struct {
	mock.Mock
}

func (m *MockLoggerServiceTestify) Info(ctx context.Context, message string, args ...interface{}) {
	m.Called(ctx, message, args)
}

func (m *MockLoggerServiceTestify) Error(ctx context.Context, message string, args ...interface{}) {
	m.Called(ctx, message, args)
}

func (m *MockLoggerServiceTestify) Warn(ctx context.Context, message string, args ...interface{}) {
	m.Called(ctx, message, args)
}
