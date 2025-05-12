# Kafka Implementation

This package provides a robust, production-ready Kafka implementation with support for various delivery semantics, consumer modes, and configuration options. The implementation uses ZooKeeper for broker coordination and cluster management.

## Features

### 1. Delivery Semantics
- **At-Least-Once** (Default)
  - Guarantees no message loss
  - May deliver messages more than once
  - Suitable for most use cases
  - Includes DLQ and retry policies

- **Exactly-Once**
  - Guarantees exactly-once delivery
  - Uses transactions and idempotence
  - Suitable for critical operations (payments, audit trails)
  - Includes deduplication

### 2. Consumer Modes
- **Latest Offset** (Default)
  - Only reads new messages after joining
  - Suitable for real-time processing
  - No historical message processing

- **Earliest Offset**
  - Reads all unprocessed messages
  - Suitable for batch processing
  - Ensures no message loss

### 3. Consumer Configuration
```go
type ConsumerConfig struct {
    MaxWait            time.Duration // Max wait for batch
    ReadBackoffMin     time.Duration // Min backoff between reads
    ReadBackoffMax     time.Duration // Max backoff between reads
    CommitInterval     time.Duration // Offset commit interval
    HeartbeatInterval  time.Duration // Consumer group heartbeat
    SessionTimeout     time.Duration // Consumer failure detection
    RebalanceTimeout   time.Duration // Group rebalance timeout
    RetentionTime      time.Duration // Message retention
    MaxAttempts        int          // Max read attempts
    IsolationLevel     string       // Read committed/uncommitted
    AutoOffsetReset    string       // Offset reset behavior
}
```

### 4. Topic Configuration
Default topic settings include:
- 7-day message retention
- 1GB segment size
- 1MB max message size
- Producer-controlled compression
- Proper cleanup policies
- Timestamp handling
- Message format version

## Usage Examples

### 1. Basic Producer Setup
```go
cfg := NewDefaultKafkaConfig(brokers, topic)
producer := NewProducer(cfg)
defer producer.Close()

err := producer.PublishMessage(ctx, []byte("message"))
```

### 2. Exactly-Once Producer
```go
cfg := NewDefaultKafkaConfig(brokers, topic).
    WithExactlyOnce().
    WithCustomExactlyOnce(ExactlyOnceConfig{
        TransactionTimeoutMs: 30000,
        EnableDeduplication: true,
    })
producer := NewProducer(cfg)
```

### 3. Consumer Setup
```go
// Basic consumer
cfg := NewDefaultKafkaConfig(brokers, topic).
    WithLatestOffset().
    WithConsumerConfig(DefaultConsumerConfig())

// Custom consumer
cfg := NewDefaultKafkaConfig(brokers, topic).
    WithEarliestOffset().
    WithConsumerConfig(ConsumerConfig{
        MaxWait: 1 * time.Second,
        CommitInterval: 10 * time.Second,
        MaxAttempts: 5,
    })
```

### 4. Topic Management
```go
// Create topic with defaults
admin := NewAdmin(cfg)
err := admin.CreateTopic(ctx, "my-topic", 0, 0, nil)

// Custom topic configuration
customConfigs := map[string]string{
    "retention.ms": "86400000",     // 1 day retention
    "max.message.bytes": "2097152", // 2MB messages
}
err = admin.CreateTopic(ctx, "my-topic", 6, 3, customConfigs)
```

## Best Practices

### 1. Producer Configuration
- Use batching for better throughput
- Configure appropriate retry policies
- Set proper timeouts
- Use compression for large messages

### 2. Consumer Configuration
- Set appropriate commit intervals
- Configure proper heartbeat intervals
- Use appropriate isolation levels
- Set proper session timeouts

### 3. Topic Configuration
- Set appropriate retention periods
- Configure proper segment sizes
- Set appropriate replication factor
- Configure proper cleanup policies

### 4. Error Handling
- Always check for errors
- Implement proper retry logic
- Use DLQ for failed messages
- Monitor consumer lag

## Performance Considerations

### 1. Producer Performance
- Use batch processing
- Configure appropriate batch size
- Set proper compression
- Use async publishing when possible

### 2. Consumer Performance
- Use appropriate batch sizes
- Configure proper fetch sizes
- Set appropriate timeouts
- Use proper backoff strategies

### 3. Topic Performance
- Set appropriate partition count
- Configure proper replication factor
- Set appropriate segment sizes
- Configure proper retention

## Monitoring and Maintenance

### 1. Key Metrics to Monitor
- Producer throughput
- Consumer lag
- Message size
- Error rates
- Retry counts
- ZooKeeper connection status
- ZooKeeper session metrics

### 2. Common Issues
- Consumer lag
- Producer backpressure
- Network issues
- Broker failures
- ZooKeeper connection issues
- ZooKeeper session timeouts

### 3. Maintenance Tasks
- Monitor topic sizes
- Check consumer groups
- Verify replication
- Monitor error rates
- Check ZooKeeper health
- Monitor ZooKeeper metrics

## Security Considerations

### 1. Authentication
- Use SASL/PLAIN
- Use SASL/SCRAM
- Use SSL/TLS
- Configure ZooKeeper authentication

### 2. Authorization
- Use ACLs
- Use RBAC
- Use proper permissions
- Configure ZooKeeper ACLs

### 3. Encryption
- Use SSL/TLS
- Use proper key management
- Use proper certificate management
- Encrypt ZooKeeper communication

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 