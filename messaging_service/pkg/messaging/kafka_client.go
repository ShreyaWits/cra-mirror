package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"messaging_service/pkg/kafka"
)

// KafkaMessagingClient is a Kafka implementation of the MessagingClient interface
type KafkaMessagingClient struct {
	brokers         []string
	defaultDelivery DeliveryOption
	useKRaftMode    bool
	activeProducers map[string]*kafka.Producer
	activeConsumers map[string]*kafka.Consumer
}

// ClientOption is a function that configures a KafkaMessagingClient
type ClientOption func(*KafkaMessagingClient)

// WithDeliveryOption sets the default delivery option
func WithDeliveryOption(delivery DeliveryOption) ClientOption {
	return func(c *KafkaMessagingClient) {
		c.defaultDelivery = delivery
	}
}

// WithKRaftMode enables KRaft mode (no Zookeeper)
func WithKRaftMode() ClientOption {
	return func(c *KafkaMessagingClient) {
		c.useKRaftMode = true
	}
}

// DefaultSubscriptionOptions returns the default subscription options
func DefaultSubscriptionOptions() SubscriptionOptions {
	return SubscriptionOptions{
		DeliveryMode:   PubSub,
		AckMode:        AutoAck,
		MaxRetries:     5,
		RetryBackoffMs: 500,
	}
}

// NewKafkaMessagingClient creates a new Kafka messaging client
func NewKafkaMessagingClient(brokers []string, options ...ClientOption) *KafkaMessagingClient {
	client := &KafkaMessagingClient{
		brokers:         brokers,
		defaultDelivery: DeliveryAtLeastOnce,
		useKRaftMode:    false,
		activeProducers: make(map[string]*kafka.Producer),
		activeConsumers: make(map[string]*kafka.Consumer),
	}

	// Apply options
	for _, option := range options {
		option(client)
	}

	return client
}

// getProducer gets or creates a producer for a topic
func (c *KafkaMessagingClient) getProducer(topic string, delivery DeliveryOption) (*kafka.Producer, error) {
	// Check if we already have a producer for this topic
	if producer, ok := c.activeProducers[topic]; ok {
		return producer, nil
	}

	// Create Kafka config
	cfg := kafka.NewDefaultKafkaConfig(c.brokers, topic)

	// Apply delivery option - we only support at-least-once and exactly-once
	switch delivery {
	case DeliveryAtLeastOnce:
		cfg = cfg.WithAtLeastOnce()
	case DeliveryExactlyOnce:
		cfg = cfg.WithExactlyOnce()
	default:
		// Default to at-least-once for any other value
		cfg = cfg.WithAtLeastOnce()
	}

	// Apply KRaft mode if enabled
	if c.useKRaftMode {
		cfg = cfg.WithKRaftMode()
	}

	// Create producer
	producer := kafka.NewProducer(cfg)
	c.activeProducers[topic] = producer

	return producer, nil
}

// Publish publishes a message to a topic
func (c *KafkaMessagingClient) Publish(ctx context.Context, topic string, payload map[string]interface{}, headers map[string]string) (string, error) {
	return c.PublishWithDelivery(ctx, topic, payload, headers, c.defaultDelivery)
}

// PublishWithDelivery publishes a message with specific delivery guarantees
func (c *KafkaMessagingClient) PublishWithDelivery(ctx context.Context, topic string, payload map[string]interface{}, headers map[string]string, delivery DeliveryOption) (string, error) {
	// Create message ID
	messageID := fmt.Sprintf("%s-%d", topic, time.Now().UnixNano())

	// Get or create producer
	producer, err := c.getProducer(topic, delivery)
	if err != nil {
		return "", fmt.Errorf("failed to get producer: %w", err)
	}

	// Add message ID to payload for tracking
	payloadWithID := make(map[string]interface{})
	for k, v := range payload {
		payloadWithID[k] = v
	}
	payloadWithID["_messageId"] = messageID

	// Marshal payload
	payloadBytes, err := json.Marshal(payloadWithID)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Convert headers to Kafka headers
	kafkaHeaders := make([]kafka.Header, 0, len(headers))
	for k, v := range headers {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}

	// Publish message with retry for at-least-once or exactly-once semantics
	if delivery == DeliveryExactlyOnce {
		// For exactly-once, use transactions
		err = producer.WriteWithRetry(ctx, []byte(messageID), payloadBytes, 5)
	} else {
		// For at-least-once (default), use retries with acks
		err = producer.WriteWithRetry(ctx, []byte(messageID), payloadBytes, 5)
	}

	if err != nil {
		return "", fmt.Errorf("failed to publish message: %w", err)
	}

	return messageID, nil
}

// Subscribe subscribes to a topic and processes messages with the handler (auto-ack)
func (c *KafkaMessagingClient) Subscribe(ctx context.Context, topic string, groupID string, handler MessageHandlerFunc) error {
	// Use default options with auto-ack
	options := DefaultSubscriptionOptions()

	return c.SubscribeWithOptions(ctx, topic, groupID, func(msgCtx *MessageContext) error {
		// Call the auto-ack handler and then automatically acknowledge the message
		err := handler(msgCtx.Context, msgCtx.Message)
		if err == nil {
			msgCtx.Ack()
		}
		return err
	}, options)
}

// SubscribeWithOptions subscribes to a topic with additional options
func (c *KafkaMessagingClient) SubscribeWithOptions(ctx context.Context, topic string, groupID string, handler MessageHandlerWithAckFunc, options SubscriptionOptions) error {
	// Create Kafka config
	cfg := kafka.NewDefaultKafkaConfig(c.brokers, topic)
	cfg.GroupID = groupID

	// If using point-to-point mode, ensure each message goes to only one consumer
	// by using a consumer group
	if options.DeliveryMode == PointToPoint {
		if groupID == "" {
			// For point-to-point, we must have a consumer group
			groupID = fmt.Sprintf("%s-p2p-%d", topic, time.Now().UnixNano())
			cfg.GroupID = groupID
		}
	}

	// Configure max retries
	if options.MaxRetries > 0 {
		// We'll set this in the RetryHandler when creating the consumer
	}

	// Always use at-least-once for consumers
	cfg = cfg.WithAtLeastOnce()

	// Apply KRaft mode if enabled
	if c.useKRaftMode {
		cfg = cfg.WithKRaftMode()
	}

	// Track acknowledgments for manual ack mode
	ackTracker := newAckTracker()

	// Create message handler
	messageHandler := func(value []byte) error {
		// Parse message
		var payload map[string]interface{}
		err := json.Unmarshal(value, &payload)
		if err != nil {
			return kafka.NewConsumerError(
				err,
				true, // Permanent error
				"Failed to parse message",
			)
		}

		// Extract message ID if present
		var messageID string
		if id, ok := payload["_messageId"]; ok {
			messageID, _ = id.(string)
			delete(payload, "_messageId") // Remove internal field
		} else {
			messageID = fmt.Sprintf("%s-%d", topic, time.Now().UnixNano())
		}

		// Create message object
		msg := Message{
			ID:        messageID,
			Topic:     topic,
			Payload:   payload,
			Timestamp: time.Now(),
		}

		// Create message context for manual ack
		msgCtx := &MessageContext{
			Message: msg,
			Context: ctx,
		}

		// Set up acknowledgment function if using manual ack mode
		if options.AckMode == ManualAck {
			// Register this message for manual acknowledgment
			ackID := ackTracker.register(messageID)

			// Set up the ack function
			msgCtx.ackFunc = func() error {
				return ackTracker.acknowledge(ackID)
			}

			// Call handler without auto-ack
			err = handler(msgCtx)

			// If message was not acknowledged, treat as error
			if !ackTracker.isAcknowledged(ackID) && err == nil {
				return fmt.Errorf("message was not acknowledged")
			}
		} else {
			// Auto-ack mode - just call the handler, ack happens in the wrapper
			err = handler(msgCtx)
		}

		if err != nil {
			return kafka.NewConsumerError(
				err,
				false, // Retriable error by default
				"Handler failed",
			)
		}

		return nil
	}

	// Create consumer with configured retries
	consumer := kafka.NewConsumer(cfg, messageHandler)

	// Override the default retry handler if specified
	if options.MaxRetries > 0 && options.RetryBackoffMs > 0 {
		consumer.RetryHandler = kafka.NewRetryHandler(
			options.MaxRetries,
			time.Duration(options.RetryBackoffMs)*time.Millisecond,
		)
	}

	c.activeConsumers[topic+":"+groupID] = consumer

	// Start consumer
	go func() {
		consumer.Start(ctx)
	}()

	return nil
}

// Close closes the messaging client
func (c *KafkaMessagingClient) Close() error {
	// Close all producers
	for _, producer := range c.activeProducers {
		if err := producer.Close(); err != nil {
			return err
		}
	}

	return nil
}

// ackTracker tracks message acknowledgments for manual ack mode
type ackTracker struct {
	mu    sync.Mutex
	acks  map[string]bool
	count int
}

// newAckTracker creates a new acknowledgment tracker
func newAckTracker() *ackTracker {
	return &ackTracker{
		acks: make(map[string]bool),
	}
}

// register registers a new message for acknowledgment tracking
func (a *ackTracker) register(messageID string) string {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.count++
	ackID := fmt.Sprintf("%s-%d", messageID, a.count)
	a.acks[ackID] = false

	return ackID
}

// acknowledge marks a message as acknowledged
func (a *ackTracker) acknowledge(ackID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.acks[ackID]; !exists {
		return fmt.Errorf("unknown acknowledgment ID: %s", ackID)
	}

	a.acks[ackID] = true
	return nil
}

// isAcknowledged checks if a message has been acknowledged
func (a *ackTracker) isAcknowledged(ackID string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	acked, exists := a.acks[ackID]
	return exists && acked
}
