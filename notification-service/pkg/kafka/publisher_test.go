package kafka

import (
	"context"
	"errors"
	"notification-service/pkg/config"
	kafka_mock "notification-service/pkg/kafka/mock"
	"notification-service/pkg/logger"
	"sync"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
)

const testTopic = "test_notification_topic"

func NewKafkaProducer(writer KafkaWriter) *KafkaPublisher {
	LOKI_URL := config.GetEnv("LOKI_URL", "http://localhost:5000")
	// Initialize logger for tests
	logger.InitLogger(LOKI_URL)
	return &KafkaPublisher{
		writer: writer,
	}
}

func TestInitKafkaPublisher(t *testing.T) {
	publisher := NewKafkaProducer(&kafka_mock.MockKafkaWriter{
		Messages: []kafka.Message{},
		Mu:       sync.Mutex{},
		Closed:   false,
		Err:      nil,
	})
	if publisher == nil {
		t.Fatal("Expected publisher to be initialized, got nil")
	}
	defer publisher.Close()
}

func TestPublishSuccess(t *testing.T) {
	publisher := NewKafkaProducer(&kafka_mock.MockKafkaWriter{
		Messages: []kafka.Message{},
		Mu:       sync.Mutex{},
		Closed:   false,
		Err:      nil,
	})
	defer publisher.Close()

	err := publisher.Publish(context.Background(), testTopic, []byte("key1"), []byte("test message"))
	if err != nil {
		t.Errorf("Expected message to be published, but got error: %v", err)
	}
}

func TestPublishToInvalidBroker(t *testing.T) {
	publisher := NewKafkaProducer(&kafka_mock.MockKafkaWriter{
		Messages: []kafka.Message{
			kafka.Message{
				Topic: testTopic,
			},
		},
		Mu:     sync.Mutex{},
		Closed: false,
		Err:    errors.New("failed to connect to broker"),
	})
	defer publisher.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := publisher.Publish(ctx, testTopic, []byte("key2"), []byte("test message"))
	if err == nil {
		t.Error("Expected error due to invalid broker, got nil")
	}
}

func TestKafkaPublisher_Close(t *testing.T) {
	publisher := NewKafkaProducer(&kafka_mock.MockKafkaWriter{
		Messages: []kafka.Message{},
		Mu:       sync.Mutex{},
		Closed:   false,
		Err:      nil,
	})
	publisher.Close()

	// Try closing again to see if it's safe (Kafka writer's Close is idempotent)
	publisher.Close()
}
