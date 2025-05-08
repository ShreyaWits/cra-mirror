package kafka

import (
	"context"
	"fmt"
	config "notification-service/internal/configs"
	kafka_mock "notification-service/pkg/kafka/mock"
	"notification-service/pkg/logger"
	"sync"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func NewKafkaConsumerWithReader(reader KafkaReader) *KafkaConsumer {
	LOKI_URL := config.GetEnv("LOKI_URL", "http://localhost:5000")
	// Initialize logger for tests
	logger.InitLogger(LOKI_URL)
	return &KafkaConsumer{
		reader:        reader,
		topicHandlers: make(map[string]func(msg kafka.Message)),
	}
}

func TestInitKafkaConsumer(t *testing.T) {

	// Initialize a mock Kafka reader
	mockReader := &kafka_mock.MockKafkaReader{
		Messages: []kafka.Message{
			kafka.Message{
				Topic: "my-topic",
				Value: []byte("test-message"),
			},
		},
		Index:  0,
		Mu:     sync.Mutex{},
		Closed: false,
	}

	// Initialize the Kafka consumer

	kc := NewKafkaConsumerWithReader(mockReader)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	called := make(chan bool, 1)

	// Register handler
	kc.Consume("my-topic", func(msg kafka.Message) {
		assert.Equal(t, "test-message", string(msg.Value))
		called <- true
	})

	// Act
	kc.Start(ctx)

	select {
	case <-called:
		// success
	case <-time.After(1 * time.Second):
		t.Fatal("Handler was not called")
	}

	// Assert close
	err := kc.Close()
	require.NoError(t, err)
}

func TestConsume(t *testing.T) {

	// Initialize a mock Kafka reader
	mockReader := &kafka_mock.MockKafkaReader{
		Messages: []kafka.Message{
			kafka.Message{
				Topic: "my-topic",
				Value: []byte("test-message"),
			},
		},
		Index:  0,
		Mu:     sync.Mutex{},
		Closed: false,
	}

	// Initialize the Kafka consumer

	consumer := NewKafkaConsumerWithReader(mockReader)
	defer consumer.Close()

	// Define a handler function
	handler := func(msg kafka.Message) {
		fmt.Printf("Message received: %s\n", string(msg.Value))
	}

	// Consume the topic
	consumer.Consume("test_topic", handler)

	// Check if the handler is registered correctly
	if _, exists := consumer.topicHandlers["test_topic"]; !exists {
		t.Error("Handler is not registered")
	}

	// Check if the reader is not closed
	if consumer.reader == nil {
		t.Error("Reader is closed")
	}
}

func TestStart(t *testing.T) {
	mockReader := &kafka_mock.MockKafkaReader{
		Messages: []kafka.Message{
			{
				Topic: "test-topic",
				Value: []byte("start-test-message"),
			},
		},
		Index:  0,
		Mu:     sync.Mutex{},
		Closed: false,
	}

	consumer := NewKafkaConsumerWithReader(mockReader)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	called := make(chan bool, 1)

	consumer.Consume("test-topic", func(msg kafka.Message) {
		assert.Equal(t, "start-test-message", string(msg.Value))
		called <- true
	})

	go consumer.Start(ctx)

	select {
	case <-called:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Handler was not called in Start()")
	}
}

func TestClose(t *testing.T) {
	mockReader := &kafka_mock.MockKafkaReader{
		Messages: []kafka.Message{},
		Index:    0,
		Mu:       sync.Mutex{},
		Closed:   false,
	}

	consumer := NewKafkaConsumerWithReader(mockReader)

	err := consumer.Close()
	require.NoError(t, err)

	if !mockReader.Closed {
		t.Error("Expected reader to be closed, but it wasn't")
	}
}
