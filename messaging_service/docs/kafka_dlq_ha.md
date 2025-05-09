# Kafka Dead Letter Queue (DLQ) and High Availability Implementation

This document outlines the implementation of Kafka Dead Letter Queues (DLQ) and High Availability (HA) features in our messaging service.

## Overview

Our Kafka implementation follows industry best practices for event-driven architecture (EDA):

1. **Dead Letter Queues (DLQ)**: Messages that repeatedly fail processing are sent to topic-specific DLQ topics
2. **Configurable Delivery Semantics**: Support for at-least-once (default) and exactly-once delivery
3. **High Availability**: Topic replication, minimum in-sync replicas, and unclean leader election prevention
4. **Exponential Backoff**: Smart retry mechanism for transient failures
5. **Leader Election**: Monitoring and management of partition leaders
6. **KRaft Mode Support**: Modern deployment without Zookeeper dependency
7. **Messaging Client Abstraction**: Unified interface for messaging operations
8. **Message Delivery Patterns**: Support for both publish/subscribe and point-to-point messaging
9. **Flexible Acknowledgment**: Support for both automatic and manual acknowledgment

## Dead Letter Queue (DLQ)

### Implementation

DLQs are implemented as separate Kafka topics with naming convention `<original_topic>.dlq`. For example, the DLQ for the `gpc` topic is `gpc.dlq`.

Key features:
- **Topic-specific configuration**: Each source topic can have its own DLQ configuration
- **Extended retention**: DLQ topics have 30-day retention (vs 7 days for regular topics)
- **Retry mechanism**: Configurable retry attempts with exponential backoff before sending to DLQ
- **Error metadata**: DLQ messages include headers with error context and original source
- **Permanent vs. Transient errors**: Ability to distinguish between permanent errors (send to DLQ immediately) and transient errors (retry first)

### Usage

Messages are sent to the DLQ in two scenarios:
1. After exceeding the maximum retry attempts
2. When encountering a permanent error (e.g., deserialization failure)

Messages in the DLQ include the following metadata:
- Original message content
- Source topic
- Failure reason
- Timestamp of failure
- Message ID for tracking

## Delivery Semantics

We support two key delivery guarantees:

### At-Least-Once (Default)
- Default delivery mode with DLQ and retry policies for resilience
- Messages are never lost but may be delivered more than once
- Consumers commit offsets only after successful processing
- Producers require acks from all replicas before considering messages sent
- Multiple delivery attempts on producer errors with exponential backoff
- Applications must be designed to handle duplicate messages

```go
// Default behavior uses at-least-once semantics
client.Publish(ctx, "my-topic", payload, headers)
```

### Exactly-Once
- Should only be used where duplicate side effects are catastrophic (e.g., payments, audit trails)
- Messages are delivered exactly once (with idempotent processing)
- Requires transaction support and idempotent producers
- Consumer needs to track processed messages to deduplicate
- Highest overhead but strongest guarantee

```go
// Explicitly request exactly-once semantics
client.PublishWithDelivery(ctx, "my-topic", payload, headers, messaging.DeliveryExactlyOnce)
```

#### Exactly-Once Implementation Details

Our exactly-once implementation includes:

- **Idempotent Producers**: Enable `enable.idempotence=true` in Kafka configuration
- **Transactional API**: Uses Kafka's transactional API with unique transaction IDs
- **Transaction Management**: Begin, commit, and abort transaction functions
- **Consumer Isolation**: Configured to read only committed transactions (`read_committed`)
- **Deduplication**: Optional application-level deduplication with configurable window

Configuration options:
```go
type ExactlyOnceConfig struct {
    EnableIdempotence     bool
    EnableTransactions    bool
    TransactionTimeoutMs  int
    TransactionalIDPrefix string
    IsolationLevel        string
    EnableDeduplication   bool
    DeduplicationWindowMs int
}
```

## Messaging Patterns

Our system supports two fundamental messaging patterns:

### Publish/Subscribe (Pub/Sub)
- One message can be delivered to multiple consumers
- Each consumer group receives its own copy of each message
- Useful for broadcasting events to multiple interested parties
- Default pattern in our messaging abstraction

```go
// Subscribe to a topic with pub/sub pattern (default)
client.Subscribe(ctx, "my-topic", "consumer-group", handler)
```

### Point-to-Point (Queue)
- Each message is delivered to exactly one consumer
- Multiple consumers can form a competing consumer pattern to scale processing
- Ensures work is distributed across consumers without duplication
- Configured explicitly in our messaging abstraction

```go
// Subscribe with point-to-point pattern
options := messaging.DefaultSubscriptionOptions()
options.DeliveryMode = messaging.PointToPoint
client.SubscribeWithOptions(ctx, "my-topic", "consumer-group", handler, options)
```

## Message Acknowledgment

Our system supports two acknowledgment modes:

### Automatic Acknowledgment
- Messages are automatically acknowledged after successful processing
- Simplifies consumer code as acknowledgment is handled by the system
- Default mode in our messaging abstraction

```go
// Subscribe with automatic acknowledgment (default)
client.Subscribe(ctx, "my-topic", "consumer-group", handler)
```

### Manual Acknowledgment
- Consumers must explicitly acknowledge messages after processing
- Gives fine-grained control over when a message is considered processed
- Useful for complex workflows where processing spans multiple steps
- If a message is not acknowledged, it will be retried

```go
// Subscribe with manual acknowledgment
options := messaging.DefaultSubscriptionOptions()
options.AckMode = messaging.ManualAck
client.SubscribeWithOptions(ctx, "my-topic", "consumer-group", handlerWithAck, options)

// Handler with manual acknowledgment
func handlerWithAck(msgCtx *messaging.MessageContext) error {
    // Process the message
    // ...
    
    // Explicitly acknowledge the message when done
    return msgCtx.Ack()
}
```

## Persistent Storage

Kafka provides built-in persistent storage for messages:

- **Message Durability**: Messages are written to disk and replicated
- **Configurable Retention**: Messages can be retained based on time or space constraints
- **Replication**: Messages are replicated across multiple brokers for fault tolerance
- **Committed Offsets**: Consumer progress is tracked and persisted

These features ensure messages are not lost in transit, even in the event of broker failures or consumer crashes.

## High Availability

Our Kafka implementation ensures high availability through:

### Topic Configuration
- **Default replication factor of 3**: Each partition has 3 replicas
- **Minimum in-sync replicas of 2**: Requires at least 2 replicas to be in-sync for writes
- **Disabled unclean leader election**: Prevents out-of-sync replicas from becoming leaders
- **Multiple partitions**: Each topic has multiple partitions for parallel processing

### Leader Election
- Our code includes monitoring for leader distribution
- Checks if leaders are properly distributed across brokers
- Alerts if all partitions have the same leader (potential failure scenario)

### Monitoring
- The `create_dlq_topics.go` utility includes leader election checking
- Logs leader distribution across brokers for each topic
- Identifies potential leader election issues

## KRaft Mode Support

Our implementation supports Kafka's KRaft mode for brokers (introduced in Kafka 2.8+), eliminating the dependency on Zookeeper:

### Benefits of KRaft Mode
- **Simplified Architecture**: No separate Zookeeper ensemble required
- **Improved Scalability**: Better scaling properties for large clusters
- **Enhanced Security**: Reduced attack surface by removing Zookeeper
- **Lower Operational Overhead**: One less component to manage and monitor

### Migration Path
To migrate from Zookeeper to KRaft mode:
1. Use the `WithKRaftMode()` option when creating a messaging client
2. Ensure Kafka brokers are running in KRaft mode (2.8+)
3. For existing deployments, follow Kafka's migration guide for metadata transfer

### Configuration
```go
// Create client with KRaft mode enabled
client := messaging.NewKafkaMessagingClient(
    []string{os.Getenv("KAFKA_BROKER")},
    messaging.WithKRaftMode(),
)
```

## Messaging Client Abstraction

We provide a clean abstraction layer for interacting with Kafka that hides implementation details:

### Key Features
- **Implementation Agnostic**: Interface-based design that could work with different backends
- **Simplified API**: Clean, high-level methods for common operations
- **Configurable Delivery**: Easy selection of delivery guarantees
- **Dependency Isolation**: Services only depend on the abstraction, not Kafka directly
- **Flexible Messaging Patterns**: Support for pub/sub and point-to-point
- **Acknowledgment Control**: Both automatic and manual acknowledgment

### Usage Example

```go
// Create client
client := messaging.NewKafkaMessagingClient(
    []string{os.Getenv("KAFKA_BROKER")},
    messaging.WithDeliveryOption(messaging.DeliveryExactlyOnce),
)

// Alternatively, with fallback:
brokers := []string{}
if broker := os.Getenv("KAFKA_BROKER"); broker != "" {
    brokers = []string{broker}
} else {
    brokers = []string{"kafka:9092"} // Default fallback
}
client := messaging.NewKafkaMessagingClient(
    brokers,
    messaging.WithDeliveryOption(messaging.DeliveryExactlyOnce),
)

// Basic pub/sub with auto-ack
client.Subscribe(ctx, "my-topic", "my-group", func(ctx context.Context, msg messaging.Message) error {
    // Process message
    return nil
})

// Point-to-point with manual ack
options := messaging.DefaultSubscriptionOptions()
options.DeliveryMode = messaging.PointToPoint
options.AckMode = messaging.ManualAck

client.SubscribeWithOptions(ctx, "my-topic", "my-group", func(msgCtx *messaging.MessageContext) error {
    // Process message
    msg := msgCtx.Message
    
    // Explicitly acknowledge when done
    return msgCtx.Ack()
}, options)
```

## Testing

You can test the DLQ functionality using:

```bash
go run messaging_service/cmd/dlq_utils/create_dlq_topics.go --test
```

You can test the messaging patterns and acknowledgment modes using:

```bash
go run messaging_service/pkg/messaging/example/example.go
```

## Configuration

DLQ topics are configured in `messaging_service/pkg/kafka/dlq_config.go`. To add a new DLQ topic:

```go
var RegisteredDLQTopics = map[string]DLQTopicConfig{
    "gpc": {
        SourceTopic: "gpc",
        DLQTopic:    "gpc.dlq",
        MaxRetries:  5,
    },
    "new-topic": {
        SourceTopic: "new-topic",
        DLQTopic:    "new-topic.dlq",
        MaxRetries:  3,
    },
}
```

Subscription options are configured when subscribing:

```go
options := messaging.DefaultSubscriptionOptions()
options.DeliveryMode = messaging.PointToPoint // or PubSub
options.AckMode = messaging.ManualAck // or AutoAck
options.MaxRetries = 3
options.RetryBackoffMs = 1000

client.SubscribeWithOptions(ctx, topic, groupID, handler, options)
``` 