package messaging

import (
	"context"
	"time"
)

// Message represents a message in the messaging system
type Message struct {
	// ID is a unique identifier for the message
	ID string `json:"id"`

	// Topic is the topic the message belongs to
	Topic string `json:"topic"`

	// Payload is the message content
	Payload map[string]interface{} `json:"payload"`

	// Headers are key-value pairs associated with the message
	Headers map[string]string `json:"headers"`

	// Timestamp is when the message was created
	Timestamp time.Time `json:"timestamp"`
}

// DeliveryOption represents message delivery guarantees
type DeliveryOption int

const (
	// DeliveryAtLeastOnce ensures messages are not lost but may be delivered more than once
	// This is the default option with DLQ and retry policies for resilience
	DeliveryAtLeastOnce DeliveryOption = iota

	// DeliveryExactlyOnce ensures messages are delivered exactly once
	// Should only be used where duplicate side effects are catastrophic (e.g., payments, audit trails)
	DeliveryExactlyOnce
)

// DeliveryMode represents the message delivery pattern
type DeliveryMode int

const (
	// PubSub allows multiple subscribers to receive the same message
	PubSub DeliveryMode = iota

	// PointToPoint ensures each message is delivered to only one consumer
	PointToPoint
)

// AckMode represents the message acknowledgment mode
type AckMode int

const (
	// AutoAck automatically acknowledges messages after handler execution
	AutoAck AckMode = iota

	// ManualAck requires explicit acknowledgment by the handler
	ManualAck
)

// MessageContext provides context for a message with acknowledgment capabilities
type MessageContext struct {
	// Message is the message being processed
	Message Message

	// Context is the processing context
	Context context.Context

	// ackFunc is the function to call to acknowledge the message
	ackFunc func() error
}

// Ack acknowledges the message, indicating successful processing
func (mc *MessageContext) Ack() error {
	if mc.ackFunc != nil {
		return mc.ackFunc()
	}
	return nil
}

// MessageHandlerFunc is a function that processes messages
type MessageHandlerFunc func(ctx context.Context, msg Message) error

// MessageHandlerWithAckFunc is a function that processes messages with manual acknowledgment
type MessageHandlerWithAckFunc func(msgCtx *MessageContext) error

// MessagingClient is the interface for the messaging system
type MessagingClient interface {
	// Publish publishes a message to a topic using the default delivery option (at-least-once)
	Publish(ctx context.Context, topic string, payload map[string]interface{}, headers map[string]string) (string, error)

	// PublishWithDelivery publishes a message with specific delivery guarantees
	PublishWithDelivery(ctx context.Context, topic string, payload map[string]interface{}, headers map[string]string, delivery DeliveryOption) (string, error)

	// Subscribe subscribes to a topic and processes messages with the handler (auto-ack)
	Subscribe(ctx context.Context, topic string, groupID string, handler MessageHandlerFunc) error

	// SubscribeWithOptions subscribes to a topic with additional options
	SubscribeWithOptions(ctx context.Context, topic string, groupID string, handler MessageHandlerWithAckFunc, options SubscriptionOptions) error

	// Close closes the messaging client
	Close() error
}

// SubscriptionOptions configures a subscription
type SubscriptionOptions struct {
	// DeliveryMode specifies if messages should be delivered in pub/sub or point-to-point mode
	DeliveryMode DeliveryMode

	// AckMode specifies if messages should be auto-acknowledged or manually acknowledged
	AckMode AckMode

	// MaxRetries specifies how many times to retry processing a message before sending to DLQ
	MaxRetries int

	// RetryBackoffMs specifies the base backoff time in milliseconds for retries
	RetryBackoffMs int
}
