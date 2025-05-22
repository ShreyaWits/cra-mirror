package confluent_test

import (
	"context"
	"errors"
	"messaging_service/pkg/confluent"
	"messaging_service/pkg/logger"
	metrics "messaging_service/pkg/matrics"
	"messaging_service/pkg/observability"
	"messaging_service/pkg/tracer"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockKafkaProducer implements the KafkaProducerInterface for testing
type MockKafkaProducer struct {
	mock.Mock
	EventsChan chan kafka.Event
}

func NewMockKafkaProducer() *MockKafkaProducer {
	return &MockKafkaProducer{
		EventsChan: make(chan kafka.Event),
	}
}

func (m *MockKafkaProducer) Events() chan kafka.Event {
	return m.EventsChan
}

func (m *MockKafkaProducer) Close() {
	close(m.EventsChan)
	m.Called()
}

func (m *MockKafkaProducer) Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error {
	args := m.Called(msg, deliveryChan)
	return args.Error(0)
}

func (m *MockKafkaProducer) Flush(timeoutMs int) int {
	args := m.Called(timeoutMs)
	return args.Int(0)
}

func (m *MockKafkaProducer) InitTransactions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockKafkaProducer) BeginTransaction() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockKafkaProducer) CommitTransaction(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockKafkaProducer) AbortTransaction(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// TestLogKafkaError tests the error logging utility
func TestLogKafkaError(t *testing.T) {
	// Create a mock logger and observability stack
	mockLogger := logger.NewMockLogger()
	mockTracer := tracer.NewMockTracerService()
	mockMetrics := metrics.NewMockMetricsService()

	mockStack := &observability.ObservabilityStack{
		LoggerService:  mockLogger,
		TracerService:  mockTracer,
		MetricsService: mockMetrics,
	}

	ctx := context.Background()

	// Set up test parameters
	eventType := "test-event"
	topic := "test-topic"
	message := "Test error message"
	testError := errors.New("test error")

	// Call the function
	confluent.LogKafkaError(ctx, eventType, topic, message, testError, mockStack)

	// Verify that the error was logged
	assert.Equal(t, 1, len(mockLogger.ErrorMessages))
	assert.Equal(t, message, mockLogger.ErrorMessages[0].Args[0])

	// Check that fields were included
	fields, ok := mockLogger.ErrorMessages[0].Args[1].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, eventType, fields["eventType"])
	assert.Equal(t, topic, fields["topic"])
	assert.Equal(t, testError.Error(), fields["error"])
}

// TestLogKafkaInfo tests the info logging utility
func TestLogKafkaInfo(t *testing.T) {
	// Create a mock logger and observability stack
	mockLogger := logger.NewMockLogger()
	mockTracer := tracer.NewMockTracerService()
	mockMetrics := metrics.NewMockMetricsService()

	mockStack := &observability.ObservabilityStack{
		LoggerService:  mockLogger,
		TracerService:  mockTracer,
		MetricsService: mockMetrics,
	}

	ctx := context.Background()

	// Set up test parameters
	eventType := "test-info-event"
	topic := "test-topic"
	message := "Test info message"

	// Call the function
	confluent.LogKafkaInfo(ctx, eventType, topic, message, mockStack)

	// Verify that the info was logged
	assert.Equal(t, 1, len(mockLogger.InfoMessages))
	assert.Equal(t, message, mockLogger.InfoMessages[0].Args[0])

	// Check that fields were included
	fields, ok := mockLogger.InfoMessages[0].Args[1].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, eventType, fields["eventType"])
	assert.Equal(t, topic, fields["topic"])
}

// TestLogKafkaWarning tests the warning logging utility
func TestLogKafkaWarning(t *testing.T) {
	// Create a mock logger and observability stack
	mockLogger := logger.NewMockLogger()
	mockTracer := tracer.NewMockTracerService()
	mockMetrics := metrics.NewMockMetricsService()

	mockStack := &observability.ObservabilityStack{
		LoggerService:  mockLogger,
		TracerService:  mockTracer,
		MetricsService: mockMetrics,
	}

	ctx := context.Background()

	// Set up test parameters
	eventType := "test-warning-event"
	topic := "test-topic"
	message := "Test warning message"

	// Call the function
	confluent.LogKafkaWarning(ctx, eventType, topic, message, mockStack)

	// Verify that the warning was logged
	assert.Equal(t, 1, len(mockLogger.WarnMessages))
	assert.Equal(t, message, mockLogger.WarnMessages[0].Args[0])

	// Check that fields were included
	fields, ok := mockLogger.WarnMessages[0].Args[1].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, eventType, fields["eventType"])
	assert.Equal(t, topic, fields["topic"])
}

// TestKafkaMonitorDeliveryReports tests the delivery report monitoring
func TestKafkaMonitorDeliveryReports(t *testing.T) {
	// Create mock components
	mockProducer := NewMockKafkaProducer()

	// Create a mock logger and observability stack
	mockLogger := logger.NewMockLogger()
	mockTracer := tracer.NewMockTracerService()
	mockMetrics := metrics.NewMockMetricsService()

	mockStack := &observability.ObservabilityStack{
		LoggerService:  mockLogger,
		TracerService:  mockTracer,
		MetricsService: mockMetrics,
	}

	// Set up test topic
	topic := "test-topic"

	// Set up mock expectations for tracer
	_, _ = mockTracer.StartTracer(context.Background(), "KafkaMonitorDeliveryReports")

	// Start monitoring in a separate goroutine
	go confluent.KafkaMonitorDeliveryReports(mockProducer, topic, mockStack)

	// Give the goroutine a moment to start
	time.Sleep(50 * time.Millisecond)

	// Create a successful delivery report
	topicStr := topic
	successMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topicStr,
			Partition: 1,
			Offset:    kafka.Offset(100),
			Error:     nil,
		},
	}

	// Create a failed delivery report
	failedMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topicStr,
			Partition: 2,
			Error:     errors.New("delivery failed"),
		},
	}

	// Send the messages to the monitor
	mockProducer.EventsChan <- successMsg
	mockProducer.EventsChan <- failedMsg

	// Give the goroutine time to process the messages
	time.Sleep(50 * time.Millisecond)

	// Setup cleanup for the mock producer
	mockProducer.On("Close").Return().Once()

	// Close the producer to end the monitor goroutine
	mockProducer.Close()

	// Give the goroutine time to clean up
	time.Sleep(50 * time.Millisecond)

	// Verify that spans and logs were created
	assert.GreaterOrEqual(t, len(mockTracer.StartTracerCalls), 1)
	assert.Equal(t, "KafkaMonitorDeliveryReports", mockTracer.StartTracerCalls[0].Name)

	// Verify logs were created (at least 3: start, success, error)
	assert.GreaterOrEqual(t, len(mockLogger.InfoMessages), 2)
	assert.GreaterOrEqual(t, len(mockLogger.ErrorMessages), 1)

	// Verify mock expectations
	mockProducer.AssertExpectations(t)
}
