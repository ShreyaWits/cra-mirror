package kafka_mock

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// MockKafkaConsumer is a mock implementation of KafkaConsumerInterface for testing.
type MockKafkaReader struct {
	Messages []kafka.Message
	Index    int
	Mu       sync.Mutex
	Closed   bool
}

func (r *MockKafkaReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	if r.Closed {
		return kafka.Message{}, errors.New("reader closed")
	}

	if r.Index >= len(r.Messages) {
		select {
		case <-ctx.Done():
			return kafka.Message{}, ctx.Err()
		case <-time.After(10 * time.Millisecond):
			return kafka.Message{}, errors.New("no more messages")
		}
	}

	msg := r.Messages[r.Index]
	r.Index++
	return msg, nil
}

func (r *MockKafkaReader) Close() error {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	r.Closed = true
	return nil
}
