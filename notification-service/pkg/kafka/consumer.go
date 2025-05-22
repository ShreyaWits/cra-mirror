// pkg/kafka/consumer.go
package kafka

import (
	"context"
	"errors" // Import errors package
	"fmt"
	"log"
	"notification-service/pkg/logger" // Assuming logger is in pkg/logger
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"
)

// KafkaConsumer wraps the underlying Kafka consumer client
type KafkaConsumer struct {
	reader        KafkaReader // changed from *kafka.Reader
	topicHandlers map[string]func(msg kafka.Message)
	started       bool
}

// InitKafkaConsumer initializes and returns a KafkaConsumer.
// In test environments, an optional mock reader can be provided (though we'll handle mock injection differently now).
func InitKafkaConsumer(brokers []string, topics []string) (*KafkaConsumer, error) {

	log.Println("Initializing Kafka reader...", brokers, topics)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     "notification-service-group",
		MinBytes:    1,
		MaxBytes:    10e6,
		GroupTopics: topics,
	})

	return &KafkaConsumer{
		reader:        reader, // kafka.Reader implements KafkaReader
		topicHandlers: make(map[string]func(msg kafka.Message)),
	}, nil
}

// Consume registers a handler function for a specific Kafka topic.
func (kc *KafkaConsumer) Consume(topic string, handler func(msg kafka.Message)) {
	if kc.started {
		// Using logger.Log.Fatalf will exit the program.
		// Consider returning an error or panicking only in truly unrecoverable situations.
		// For this example, we'll keep it as is but be mindful in a real application.
		logger.Log.Fatalf("Cannot register new topic after consumer has started")
		panic("Cannot register new topic after consumer has started")
	}

	kc.topicHandlers[topic] = handler
	logger.Log.Printf("Handler registered for topic: %s", topic)
}

// Start begins consuming messages with the given handler function
func (kc *KafkaConsumer) Start(ctx context.Context) {
	if kc.started {
		logger.Log.Println("Consumer already started")
		return
	}

	logger.Log.Println("Starting Kafka consumer...")

	kc.started = true

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// We should use the context provided to ReadMessage to handle cancellation
	// The loop structure needs adjustment to properly use the context for shutdown
	// and also listen for OS signals.

	go func() { // Run consumption in a goroutine to not block
		for {
			select {
			case <-ctx.Done():
				logger.Log.Println("Context cancelled. Stopping consumer read loop.")
				return // Exit the goroutine
			case sig := <-sigChan:
				logger.Log.Printf("Caught signal %v: stopping consumer read loop\n", sig)
				return // Exit the goroutine
			default:
				// ReadMessage should ideally block until a message is received or context is cancelled
				msg, err := kc.reader.ReadMessage(ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						logger.Log.Println("Context cancelled during ReadMessage.")
						return // Context cancelled, exit goroutine
					}
					logger.Log.Printf("Error reading message: %v", err)
					// Depending on the error, you might want to continue or break
					continue
				}

				topic := msg.Topic
				if handler, exists := kc.topicHandlers[topic]; exists {
					// Handle the message - perhaps in a new goroutine if processing is long
					go handler(msg)
				} else {
					logger.Log.Printf("No handler for topic: %s", topic)
				}
			}
		}
	}()
}

// Close gracefully shuts down the consumer
func (kc *KafkaConsumer) Close() error { // Modified to return error
	logger.Log.Println("Closing Kafka consumer...")
	if err := kc.reader.Close(); err != nil {
		logger.Log.Printf("Error closing reader: %v", err)
		return fmt.Errorf("failed to close Kafka reader: %w", err) // Return error
	}
	logger.Log.Println("Kafka consumer closed.")
	return nil // Return nil on success
}
