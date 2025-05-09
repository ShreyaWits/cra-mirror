package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"messaging_service/pkg/messaging"
)

func main() {
	fmt.Println("Starting messaging example...")

	// Parse Kafka brokers from env or use default
	brokers := []string{}

	// First check KAFKA_BROKER (singular)
	if broker := os.Getenv("KAFKA_BROKER"); broker != "" {
		brokers = []string{broker}
	} else if brokersEnv := os.Getenv("KAFKA_BROKERS"); brokersEnv != "" {
		// Fall back to KAFKA_BROKERS (plural)
		brokers = strings.Split(brokersEnv, ",")
	} else {
		// Use default only if no env vars are set
		brokers = []string{"kafka:9092"}
	}

	// Create messaging client with defaults (at-least-once delivery)
	client := messaging.NewKafkaMessagingClient(
		brokers,
	)
	defer client.Close()

	// Create a context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		log.Println("Shutting down...")
		cancel()
	}()

	// Example topics
	topic := "example-topic"
	dlqTopic := "example-topic.dlq"

	// Example 1: Basic publishing with default at-least-once delivery
	ctx = context.Background()
	payload := map[string]interface{}{
		"message": "Hello world",
		"time":    time.Now().Format(time.RFC3339),
	}
	msgID, err := client.Publish(ctx, topic, payload, map[string]string{
		"example-header": "value",
	})
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}
	fmt.Printf("Published message %s with at-least-once delivery (default)\n", msgID)

	// Example 2: Publishing with exactly-once delivery semantics
	exactlyOncePayload := map[string]interface{}{
		"message": "Exactly once message",
		"time":    time.Now().Format(time.RFC3339),
	}
	exactlyOnceMsgID, err := client.PublishWithDelivery(ctx, topic, exactlyOncePayload,
		map[string]string{"delivery": "exactly-once"},
		messaging.DeliveryExactlyOnce)
	if err != nil {
		log.Fatalf("Failed to publish exactly-once message: %v", err)
	}
	fmt.Printf("Published message %s with exactly-once delivery\n", exactlyOnceMsgID)

	// Example 3: Pub/Sub pattern with auto acknowledgment (default)
	go func() {
		err := client.Subscribe(ctx, topic, "example-consumer-group", func(ctx context.Context, msg messaging.Message) error {
			fmt.Printf("Received message in pub/sub mode (auto-ack): %v\n", msg.Payload["message"])
			fmt.Printf("Headers: %v\n", msg.Headers)
			return nil // Auto-acknowledge
		})
		if err != nil {
			log.Printf("Subscription error: %v", err)
		}
	}()

	// Example 4: Point-to-point pattern with manual acknowledgment
	go func() {
		options := messaging.DefaultSubscriptionOptions()
		options.DeliveryMode = messaging.PointToPoint
		options.AckMode = messaging.ManualAck

		err := client.SubscribeWithOptions(ctx, topic, "example-p2p-group", func(msgCtx *messaging.MessageContext) error {
			msg := msgCtx.Message
			fmt.Printf("Received message in point-to-point mode (manual-ack): %v\n", msg.Payload["message"])

			// Simulate processing
			time.Sleep(100 * time.Millisecond)

			// Manually acknowledge
			return msgCtx.Ack()
		}, options)

		if err != nil {
			log.Printf("P2P subscription error: %v", err)
		}
	}()

	// Example 5: DLQ processing - consumer for the DLQ topic
	go func() {
		err := client.Subscribe(ctx, dlqTopic, "dlq-consumer", func(ctx context.Context, msg messaging.Message) error {
			fmt.Printf("DLQ message: %v\n", msg.Payload["message"])
			fmt.Printf("Original error: %s\n", msg.Headers["error-reason"])
			fmt.Printf("Failed attempts: %s\n", msg.Headers["retry-count"])
			fmt.Printf("Original topic: %s\n", msg.Headers["source-topic"])
			return nil
		})
		if err != nil {
			log.Printf("DLQ subscription error: %v", err)
		}
	}()

	// Example 6: Demonstrate message failure that will eventually go to DLQ
	go func() {
		// Consumer that fails processing messages
		failingOptions := messaging.DefaultSubscriptionOptions()
		failingOptions.MaxRetries = 3 // After 3 failures, message goes to DLQ

		err := client.SubscribeWithOptions(ctx, topic, "failing-consumer", func(msgCtx *messaging.MessageContext) error {
			fmt.Printf("Processing message, but will fail: %v\n", msgCtx.Message.Payload["message"])
			return fmt.Errorf("simulated processing error")
		}, failingOptions)

		if err != nil {
			log.Printf("Failing consumer subscription error: %v", err)
		}
	}()

	// Allow time for processing
	time.Sleep(30 * time.Second)
	fmt.Println("Example complete")
}

// DefaultSubscriptionOptions returns default subscription options
// This is a placeholder function since the actual implementation would be in the messaging package
func DefaultSubscriptionOptions() messaging.SubscriptionOptions {
	return messaging.SubscriptionOptions{
		DeliveryMode:   messaging.PubSub,
		AckMode:        messaging.AutoAck,
		MaxRetries:     3,
		RetryBackoffMs: 1000,
	}
}
