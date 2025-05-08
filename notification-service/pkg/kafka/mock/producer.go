package kafka_mock

import (
	"context"
	"errors"
	"sync"

	"github.com/segmentio/kafka-go"
)

// MockKafkaWriter simulates a Kafka writer for testing purposes.
type MockKafkaWriter struct {
	Messages []kafka.Message // Captures messages sent
	Mu       sync.Mutex      // Ensures thread safety
	Closed   bool            // Tracks if the writer is closed
	Err      error           // Optional error to simulate failure
}

// WriteMessages simulates publishing messages to Kafka.
func (w *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	w.Mu.Lock()
	defer w.Mu.Unlock()

	if w.Closed {
		return errors.New("writer closed")
	}
	if w.Err != nil {
		return w.Err // Simulate write error
	}

	w.Messages = append(w.Messages, msgs...)
	return nil
}

// Close marks the writer as closed.
func (w *MockKafkaWriter) Close() error {
	w.Mu.Lock()
	defer w.Mu.Unlock()
	w.Closed = true
	return nil
}
