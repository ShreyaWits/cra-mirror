# Kafka Messaging Service API Documentation

## Overview
This service provides a high-performance Kafka messaging implementation with support for topic creation, message publishing, and message subscription. The service uses the segmentio/kafka-go client library.

## Configuration Types

The service uses three types of configurations:

1. **Kafka-native configurations**: Direct Kafka broker settings
2. **Client-library configurations**: segmentio/kafka-go specific settings
3. **Service-level configurations**: Our abstraction layer settings

## API Endpoints

### 1. Create Topic
Creates a new Kafka topic with specified configuration.

**Request:**
```json
{
  "topic": "my-temporary-topic",
  "num_partitions": 1,
  "replication_factor": 1,
  "config": {
    "retention.ms": "3000",
    "cleanup.policy": "delete"
  }
}
```

**Fields:**
- `topic` (string, required): Name of the topic to create
- `num_partitions` (int, required): Number of partitions
- `replication_factor` (int, required): Replication factor
- `config` (map, optional): Kafka-native topic configuration
  - `retention.ms` (string): Message retention time in milliseconds
  - `cleanup.policy` (string): Topic cleanup policy ("delete" or "compact")

**Response:**
```json
{
  "status": "success",
  "message": "Topic my-temporary-topic created successfully",
  "validation_errors": []
}
```

### 2. Publish Message
Publishes a message to a Kafka topic.

**Request:**
```json
{
  "topic": "my-temporary-topic",
  "value": {
    "action": "login123",
    "timestamp": "2023-10-01T12:00:00Z"
  },
  "producer_config": {
    "enable_idempotence": true,
    "delivery_semantics": "at-least-once",
    "retries": 3,
    "retry_backoff_ms": 100,
    "timeout_ms": 30000,
    "enable_compression": true,
    "compression_type": "gzip",
    "dlq_topic": "my-temporary-topic-dlq",
    "max_retries": 3,
    "retry_delay_ms": 1000
  },
  "headers": {
    "Content-Type": "application/json",
    "Source": "web-app"
  },
  "key": "user123"
}
```

**Fields:**
- `topic` (string, required): Topic to publish to
- `value` (map, required): Message content
- `producer_config` (object, optional): Producer configuration
  - **Kafka-native settings:**
    - `enable_idempotence` (boolean): Enable idempotent producer
    - `delivery_semantics` (string): "at-least-once" (default) or "exactly-once"
  - **Client-library settings:**
    - `retries` (int): Number of retries (kafka-go specific)
    - `retry_backoff_ms` (int): Backoff time between retries
    - `timeout_ms` (int): Overall operation timeout
    - `enable_compression` (boolean): Enable message compression
    - `compression_type` (string): "gzip", "snappy", "lz4", "zstd"
  - **Service-level settings:**
    - `dlq_topic` (string): Dead Letter Queue topic name
    - `max_retries` (int): Maximum retries before DLQ
    - `retry_delay_ms` (int): Delay between retries
- `headers` (map, optional): Message headers
- `key` (string, optional): Message key for partitioning (uses hash-based partitioning)

**Response:**
```json
{
  "status": "success",
  "message": "Message published successfully",
  "partition": 0,
  "offset": 123,
  "timestamp": 1696156800000
}
```

### 3. Subscribe to Topic
Subscribes to a topic and receives messages via gRPC stream.

**Request:**
```json
{
  "topic": "my-temporary-topic",
  "group_id": "my-consumer-group",
  "consumer_config": {
    "isolation_level": "read_committed",
    "auto_offset_reset": "latest",
    "max_wait_ms": 5000,
    "read_backoff_min_ms": 100,
    "read_backoff_max_ms": 1000,
    "commit_interval_ms": 5000,
    "heartbeat_interval_ms": 3000,
    "session_timeout_ms": 30000,
    "rebalance_timeout_ms": 60000,
    "max_attempts": 3
  }
}
```

**Fields:**
- `topic` (string, required): Topic to subscribe to
- `group_id` (string, required): Consumer group ID
- `consumer_config` (object, optional): Consumer configuration
  - **Kafka-native settings:**
    - `isolation_level` (string): "read_committed" or "read_uncommitted" (default)
    - `auto_offset_reset` (string): "latest" or "earliest"
  - **Client-library settings:**
    - `max_wait_ms` (int): Maximum time to wait for messages
    - `read_backoff_min_ms` (int): Minimum backoff between reads
    - `read_backoff_max_ms` (int): Maximum backoff between reads
    - `commit_interval_ms` (int): Offset commit interval
    - `heartbeat_interval_ms` (int): Consumer group heartbeat interval
    - `session_timeout_ms` (int): Session timeout
    - `rebalance_timeout_ms` (int): Rebalance timeout
    - `max_attempts` (int): Maximum read attempts

**Stream Response:**
```json
{
  "value": {
    "action": "login123",
    "timestamp": "2023-10-01T12:00:00Z"
  },
  "timestamp": 1696156800000,
  "key": "user123",
  "headers": {
    "Content-Type": "application/json",
    "Source": "web-app"
  }
}
```

## Message Partitioning

The service uses hash-based partitioning:
- Messages with the same key go to the same partition
- Keys are hashed using a consistent hashing algorithm
- Ensures message ordering within a partition
- No custom partitioning strategies are supported

## Dead Letter Queue (DLQ)

Failed messages are handled as follows:
1. Message processing fails
2. Retry up to `max_retries` times with `retry_delay_ms` between attempts
3. If all retries fail, message is sent to the DLQ topic
4. DLQ topic name is specified in `dlq_topic`
5. Original message headers are preserved with additional error information

## Delivery Semantics

1. **At-Least-Once Delivery** (Default)
   - Messages are guaranteed to be delivered at least once
   - May result in duplicate messages
   - Use when message ordering is important
   - Set `delivery_semantics` to "at-least-once"

2. **Exactly-Once Delivery**
   - Messages are delivered exactly once
   - Requires transactional producer
   - Higher latency but stronger guarantees
   - Set `delivery_semantics` to "exactly-once"

## Consumer Modes

### Offset Reset Behavior
- `latest`: Start consuming from the latest offset
- `earliest`: Start consuming from the earliest offset

### Isolation Levels
- `read_uncommitted` (default): Reads all messages, including those from aborted transactions
- `read_committed`: Only reads messages from committed transactions

## Best Practices

1. **Topic Configuration**
   - Set appropriate retention period
   - Choose correct cleanup policy
   - Configure proper replication factor

2. **Producer Configuration**
   - Enable idempotence for critical operations
   - Configure appropriate retry settings
   - Use compression for large messages
   - Choose appropriate delivery semantics
   - Always configure DLQ for failed messages

3. **Consumer Configuration**
   - Set appropriate timeouts
   - Configure proper backoff settings
   - Choose correct isolation level
   - Set appropriate offset reset behavior

4. **Message Structure**
   - Include timestamps
   - Use meaningful keys
   - Add necessary headers

## Error Handling

1. **Producer Errors**
   - Retry on transient failures
   - Use DLQ for failed messages
   - Monitor error rates

2. **Consumer Errors**
   - Handle message processing failures
   - Implement retry logic
   - Monitor consumer lag

## Performance Considerations

1. **Message Size**
   - Keep messages under 1MB
   - Use compression for large messages
   - Batch small messages

2. **Throughput**
   - Configure appropriate batch sizes
   - Set proper timeouts
   - Monitor consumer lag

3. **Resource Usage**
   - Monitor memory usage
   - Check disk space
   - Watch network bandwidth

## Rate Limits

- Maximum message size: 1MB
- Maximum batch size: 1000 messages
- Maximum topics per cluster: 1000
- Maximum consumer groups: 1000

## Message Keys Usage

### Purpose
Message keys are used for:
1. Message partitioning
2. Message ordering
3. Message deduplication
4. Message tracking

### Key Types and Examples

1. **User-based Keys:**
```json
{
  "key": "user-123",
  "value": {
    "action": "login",
    "userId": "123"
  }
}
```

2. **Session-based Keys:**
```json
{
  "key": "session-abc-xyz",
  "value": {
    "action": "session_start",
    "sessionId": "abc-xyz"
  }
}
```

3. **Event-based Keys:**
```json
{
  "key": "event-2023-10-01-001",
  "value": {
    "eventType": "payment",
    "amount": 100.00
  }
}
```

4. **Composite Keys:**
```json
{
  "key": "user-123-order-456",
  "value": {
    "userId": "123",
    "orderId": "456",
    "action": "order_placed"
  }
}
```

## Expected Values

### Message Value Structure
```json
{
  "action": "string",      // Action type (e.g., "login", "logout", "purchase")
  "timestamp": "string",   // ISO 8601 timestamp
  "data": {               // Optional additional data
    "key1": "value1",
    "key2": "value2"
  }
}
```

### Common Action Types
- `login`
- `logout`
- `purchase`
- `refund`
- `update`
- `delete`

### Header Values
```json
{
  "Content-Type": "application/json",
  "Source": "web-app",
  "Version": "1.0",
  "Correlation-ID": "uuid-string"
}
```

## Best Practices

1. **Key Selection:**
   - Use consistent key patterns
   - Choose keys that ensure even distribution
   - Consider message ordering requirements

2. **Message Structure:**
   - Keep messages small and focused
   - Include timestamps
   - Use consistent field names

3. **Headers:**
   - Include content type
   - Add correlation IDs for tracking
   - Specify message version

4. **Error Handling:**
   - Implement retry logic
   - Use DLQ for failed messages
   - Log errors with context

## Performance Considerations

1. **Message Size:**
   - Keep messages under 1MB
   - Compress large messages
   - Use efficient serialization

2. **Throughput:**
   - Use appropriate batch sizes
   - Implement backpressure handling
   - Monitor consumer lag

3. **Retention:**
   - Set appropriate retention periods
   - Monitor disk usage
   - Clean up old messages

## Error Codes

- `400`: Bad Request
- `401`: Unauthorized
- `403`: Forbidden
- `404`: Topic Not Found
- `409`: Conflict
- `500`: Internal Server Error

## Rate Limits

- Maximum message size: 1MB
- Maximum batch size: 1000 messages
- Maximum retention period: 7 days
- Maximum topics per cluster: 1000

## Glossary

### Basic Terms
- **Topic**: A category or feed name to which messages are published. Think of it as a named channel for messages.
- **Partition**: A division of a topic that allows for parallel processing. Like chapters in a book, each partition contains a portion of the messages.
- **Message**: A unit of data that is sent through Kafka. Contains the actual information you want to transmit.
- **Key**: A unique identifier for a message that determines which partition it goes to. Like a postal code that determines which post office handles your mail.

### Consumer Terms
- **Consumer**: A client that reads messages from Kafka topics. Like a subscriber who reads a newspaper.
- **Consumer Group**: A set of consumers that work together to read messages. Like a team of readers dividing up the work.
- **Offset**: A position marker that shows where a consumer has read up to in a partition. Like a bookmark in a book.
- **Backoff**: A waiting period between retries when something fails. Like waiting a few minutes before trying to call someone again when the line is busy.
  - **Read Backoff**: The time a consumer waits between attempts to read messages when none are available.
  - **Retry Backoff**: The time a producer waits between attempts to send a message when it fails.

### Producer Terms
- **Producer**: A client that sends messages to Kafka topics. Like a publisher who writes articles for a newspaper.
- **Idempotence**: A property that ensures the same message won't be processed multiple times. Like a bank ensuring you can't withdraw the same money twice.
- **Compression**: Reducing the size of messages to save space and bandwidth. Like zipping a file before sending it.

### Configuration Terms
- **Retention**: How long messages are kept in Kafka before being deleted. Like how long a library keeps old newspapers.
- **Replication**: Making copies of data across multiple servers for reliability. Like having backup copies of important documents.
- **Isolation Level**: How strictly a consumer reads messages in relation to transactions. Like choosing between reading a draft (uncommitted) or final version (committed) of a document.

### Error Handling Terms
- **Dead Letter Queue (DLQ)**: A special topic where failed messages are sent for later processing. Like a "to be fixed" folder for problematic items.
- **Retry**: Attempting an operation again after it fails. Like trying to call someone again when they don't answer.
- **Max Attempts**: The maximum number of times to retry an operation before giving up. Like trying to call someone three times before leaving a message.

### Performance Terms
- **Throughput**: The rate at which messages are processed. Like how many cars can pass through a tunnel per hour.
- **Latency**: The time it takes to process a message. Like how long it takes for a letter to reach its destination.
- **Batch Size**: The number of messages processed together. Like how many letters a mail carrier delivers in one trip.

### Technical Terms
- **gRPC**: A modern framework for building APIs. Like a more efficient way to make phone calls between services.
- **Protocol Buffers**: A data format for serializing structured data. Like a standardized way to package messages.
- **Hash-based Partitioning**: A method of distributing messages across partitions based on their keys. Like sorting mail into different bins based on postal codes.

### Common Abbreviations
- **DLQ**: Dead Letter Queue
- **ms**: Milliseconds (1/1000 of a second)
- **KB/MB**: Kilobyte/Megabyte (units of data size)

## Best Practices

1. **Key Selection:**
   - Use consistent key patterns
   - Choose keys that ensure even distribution
   - Consider message ordering requirements

2. **Message Structure:**
   - Keep messages small and focused
   - Include timestamps
   - Use consistent field names

3. **Headers:**
   - Include content type
   - Add correlation IDs for tracking
   - Specify message version

4. **Error Handling:**
   - Implement retry logic
   - Use DLQ for failed messages
   - Log errors with context

## Performance Considerations

1. **Message Size:**
   - Keep messages under 1MB
   - Compress large messages
   - Use efficient serialization

2. **Throughput:**
   - Use appropriate batch sizes
   - Implement backpressure handling
   - Monitor consumer lag

3. **Retention:**
   - Set appropriate retention periods
   - Monitor disk usage
   - Clean up old messages

## Error Codes

- `400`: Bad Request
- `401`: Unauthorized
- `403`: Forbidden
- `404`: Topic Not Found
- `409`: Conflict
- `500`: Internal Server Error

## Rate Limits

- Maximum message size: 1MB
- Maximum batch size: 1000 messages
- Maximum retention period: 7 days
- Maximum topics per cluster: 1000 