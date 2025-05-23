package mock

import (
	"context"
	"testing"

	"messaging_service/pkg/confluent"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/mock"
)

// MockConsumerGroupListing defines a simplified consumer group listing struct for testing
type MockConsumerGroupListing struct {
	GroupID string
}

// MockConsumerGroupListings defines a simplified consumer group listings struct for testing
type MockConsumerGroupListings struct {
	Valid []MockConsumerGroupListing
}

// MockAdminClientTestify is a mock implementation of the AdminClient for testify/mock based tests
type MockAdminClientTestify struct {
	mock.Mock
}

func NewMockAdminClientTestify() *MockAdminClientTestify {
	return &MockAdminClientTestify{}
}

func (m *MockAdminClientTestify) CreateTopics(ctx context.Context, topics []confluent.TopicSpecification) ([]kafka.TopicResult, error) {
	args := m.Called(ctx, topics)
	return args.Get(0).([]kafka.TopicResult), args.Error(1)
}

func (m *MockAdminClientTestify) DeleteTopics(ctx context.Context, topics []string) ([]kafka.TopicResult, error) {
	args := m.Called(ctx, topics)
	return args.Get(0).([]kafka.TopicResult), args.Error(1)
}

func (m *MockAdminClientTestify) GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error) {
	args := m.Called(topic, allTopics, timeoutMs)
	return args.Get(0).(*kafka.Metadata), args.Error(1)
}

func (m *MockAdminClientTestify) ListConsumerGroups(ctx context.Context) (MockConsumerGroupListings, error) {
	args := m.Called(ctx)
	return args.Get(0).(MockConsumerGroupListings), args.Error(1)
}

func (m *MockAdminClientTestify) Close() {
	m.Called()
}

// MockKafkaAdminTestify is a mock implementation of the KafkaAdmin interface for tests
type MockKafkaAdminTestify struct {
	mock.Mock
}

func NewMockKafkaAdminTestify() *MockKafkaAdminTestify {
	return &MockKafkaAdminTestify{}
}

func (m *MockKafkaAdminTestify) CreateTopic(ctx context.Context, topic string, numPartitions, replicationFactor int, configs map[string]string) error {
	args := m.Called(ctx, topic, numPartitions, replicationFactor, configs)
	return args.Error(0)
}

func (m *MockKafkaAdminTestify) DeleteTopic(ctx context.Context, topic string) error {
	args := m.Called(ctx, topic)
	return args.Error(0)
}

func (m *MockKafkaAdminTestify) ListTopics(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockKafkaAdminTestify) HasActiveConsumers(ctx context.Context, topic string) (bool, error) {
	args := m.Called(ctx, topic)
	return args.Bool(0), args.Error(1)
}

func (m *MockKafkaAdminTestify) Close() error {
	args := m.Called()
	return args.Error(0)
}

// SetupMockAdmin creates a new AdminImpl with mocked dependencies using testify/mock
func SetupMockAdmin(t *testing.T) (confluent.KafkaAdmin, *MockAdminClientTestify, *MockObservabilityStackTestify) {
	mockAdmin := NewMockAdminClientTestify()
	mockObs := NewMockObservabilityStackTestify()

	// Setup common expectations
	mockObs.TracerService.On("StartTracer", mock.Anything, mock.Anything).Return(context.Background(), "span")
	mockObs.TracerService.On("StopSpan", mock.Anything).Return()

	// Instead of creating a real AdminImpl which has errors, create our mock
	admin := NewMockKafkaAdminTestify()

	return admin, mockAdmin, mockObs
}
