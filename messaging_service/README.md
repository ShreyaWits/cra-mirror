# Kafka Messaging Service

A high-performance Kafka messaging service implementation with support for topic creation, message publishing, and message subscription via gRPC.

## Features

- Topic management with configurable retention policies
- High-performance message publishing with compression
- Real-time message subscription via gRPC streams
- Dead Letter Queue (DLQ) support for failed messages
- Message retry mechanism with exponential backoff
- Comprehensive monitoring and logging

## Prerequisites

- Go 1.19 or higher
- Docker and Docker Compose
- Make (optional, for using Makefile commands)
- Protocol Buffers compiler (protoc)

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/messaging-service.git
cd messaging-service

# Install dependencies
go mod download
```

## Running the Service

### Using Docker Compose

1. **Start the service with Kafka:**
```bash
docker compose -f 'messaging_service/docker-compose.yml' up --build 'messaging_service'
```

2. **Check service status:**
```bash
docker-compose ps
```

3. **View logs:**
```bash
docker-compose logs -f
```

### Running Locally

1. **Start Kafka (if not using Docker):**
```bash
# Start Zookeeper
bin/zookeeper-server-start.sh config/zookeeper.properties

# Start Kafka
bin/kafka-server-start.sh config/server.properties
```

2. **Run the service:**
```bash
go run cmd/server/main.go
```

## How to Use the gRPC Service in Your Project

### 1. Import the Proto Package

Include the messaging service proto package in your project:

```go
import (
    pb "messaging_service/protos/messaging_service"
    "google.golang.org/grpc"
)
```

### 2. Connect to the Service

```go
// Create a gRPC connection
conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
defer conn.Close()

// Create a client
client := pb.NewMessagingServiceClient(conn)
```

### 3. Common Usage Patterns

#### Creating a Topic
```go
resp, err := client.CreateTopicV1(ctx, &pb.CreateTopicRequest{
    Topic: "my-topic",
})
if err != nil {
    log.Fatalf("Failed to create topic: %v", err)
}
log.Printf("Response: %s - %s", resp.Status, resp.Message)
```

#### Publishing a Message
```go
// Create message with map of string values
resp, err := client.PublishMessageV1(ctx, &pb.PublishRequest{
    Topic: "my-topic",
    Value: map[string]string{
        "action":    "user_login",
        "timestamp": time.Now().Format(time.RFC3339),
        "data":      "additional information",
    },
    Key: "user-123",
})
if err != nil {
    log.Fatalf("Failed to publish message: %v", err)
}
log.Printf("Response: %s - %s", resp.Status, resp.Message)
```

#### Subscribing to Messages
```go
stream, err := client.SubscribeV1(ctx, &pb.SubscribeRequest{
    Topic:   "my-topic",
    GroupId: "my-service-group",
})
if err != nil {
    log.Fatalf("Failed to subscribe: %v", err)
}

// Process incoming messages
for {
    msg, err := stream.Recv()
    if err == io.EOF {
        // Stream ended
        break
    }
    if err != nil {
        log.Fatalf("Error receiving message: %v", err)
        break
    }
    
    // Process the received message
    log.Printf("Received message: Key=%s, Timestamp=%d", msg.Key, msg.Timestamp)
    log.Printf("Message action: %s", msg.Value["action"])
    log.Printf("Message timestamp: %s", msg.Value["timestamp"])
}
```

### 4. Error Handling

Implement proper error handling:

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

resp, err := client.PublishMessageV1(ctx, req)
if err != nil {
    if st, ok := status.FromError(err); ok {
        // Handle specific gRPC status codes
        switch st.Code() {
        case codes.NotFound:
            log.Printf("Topic not found: %v", st.Message())
        case codes.InvalidArgument:
            log.Printf("Invalid request: %v", st.Message())
        default:
            log.Printf("Error: %v", st.Message())
        }
    } else {
        log.Printf("Unknown error: %v", err)
    }
    return
}
```

### 5. Connection Management

For production use, implement connection management:

```go
// Connection with timeout and retry
func createConnection() (*grpc.ClientConn, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    var conn *grpc.ClientConn
    var err error
    
    for retries := 3; retries > 0; retries-- {
        conn, err = grpc.DialContext(ctx, "localhost:9090", 
            grpc.WithInsecure(),
            grpc.WithBlock())
            
        if err == nil {
            return conn, nil
        }
        
        time.Sleep(1 * time.Second)
    }
    
    return nil, err
}
```

## API Usage Guide

### Step 1: Create a Topic

Before publishing or subscribing to messages, a topic must be created.

**Sample Code Implementation:**

```go
package main

import (
    "context"
    "log"
    "time"

    pb "messaging_service/protos/messaging_service"
    "google.golang.org/grpc"
)

func createTopic() {
    conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewMessagingServiceClient(conn)
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    resp, err := client.CreateTopicV1(ctx, &pb.CreateTopicRequest{
        Topic: "my-temporary-topic",
    })
    
    if err != nil {
        log.Fatalf("Failed to create topic: %v", err)
    }
    
    log.Printf("Response: %s - %s", resp.Status, resp.Message)
}
```

#### Topic Name Validation

When creating a topic, it is important to adhere to the following naming conventions:

- **Allowed Characters**: Topic names can contain:
  - Alphanumeric characters (a-z, A-Z, 0-9)
  - Special characters: `.`, `_`, `-`
  
- **Restrictions**:
  - Topic names must not be empty
  - Topic names must not contain spaces or special characters such as `!`, `@`, `#`, `$`, `%`, etc.
  
- **Examples of Valid Topic Names**:
  - `valid-topic`
  - `valid_topic`
  - `valid.topic`
  - `valid-topic-123`

### Step 2: Publish Messages

Publishing messages enables applications to send data to consumers.

**Sample Code Implementation:**

```go
package main

import (
    "context"
    "log"
    "time"

    pb "messaging_service/protos/messaging_service"
    "google.golang.org/grpc"
)

func publishMessage() {
    conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewMessagingServiceClient(conn)
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    resp, err := client.PublishMessageV1(ctx, &pb.PublishRequest{
        Topic: "my-temporary-topic",
        Value: map[string]string{
            "action":    "login",
            "timestamp": "2023-10-01T12:00:00Z",
        },
        Key: "user123",
    })
    
    if err != nil {
        log.Fatalf("Failed to publish message: %v", err)
    }
    
    log.Printf("Response: %s - %s", resp.Status, resp.Message)
}
```

### Message Keys Usage

The key determines which partition a message is sent to, ensuring messages with the same key are delivered in order. Keys are important for:

1. **Partitioning**: Determines the partition a message is sent to
2. **Ordering**: Maintains order of messages with the same key within a partition
3. **Deduplication**: Assists in identifying duplicate messages
4. **Tracking**: Facilitates tracing message flow through systems

Examples of meaningful keys:
- User-based: `user-123`
- Session-based: `session-abc-xyz`
- Event-based: `event-2023-10-01-001`
- Composite: `user-123-order-456`

### Step 3: Subscribe to Messages

Subscribing allows applications to receive and process messages from topics.

**Sample Code Implementation:**

```go
package main

import (
    "context"
    "io"
    "log"
    "time"

    pb "messaging_service/protos/messaging_service"
    "google.golang.org/grpc"
)

func subscribeToMessages() {
    conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewMessagingServiceClient(conn)
    
    // Context with timeout for long-lived operations
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
    defer cancel()
    
    stream, err := client.SubscribeV1(ctx, &pb.SubscribeRequest{
        Topic:   "my-temporary-topic",
        GroupId: "my-consumer-group",
    })
    
    if err != nil {
        log.Fatalf("Failed to subscribe: %v", err)
    }
    
    for {
        msg, err := stream.Recv()
        if err == io.EOF {
            // Stream ended
            break
        }
        if err != nil {
            log.Fatalf("Error receiving message: %v", err)
            break
        }
        
        // Process the received message
        log.Printf("Received message: Key=%s, Timestamp=%d", msg.Key, msg.Timestamp)
        log.Printf("Message action: %s", msg.Value["action"])
    }
}
```

#### Group ID Validation

When subscribing to messages, adhere to the following naming conventions for the `groupId`:

- **Allowed Characters**: Group IDs can contain alphanumeric characters and `.`, `_`, `-`
- **Restrictions**: No spaces or other special characters
- **Examples of Valid Group IDs**: `my-consumer-group`, `group_123`, `group.name`

#### Consumer Group Behavior

If a topic has two consumers with the same topic and same group ID, then those consumers are part of the same consumer group, and Kafka will load balance partitions among them:

- Each partition is assigned to exactly one consumer in the group
- No duplicate message delivery within a consumer group
- Rebalancing occurs when consumers join or leave the group

If you want multiple consumers to receive all messages, use different group IDs for each consumer.

## Best Practices

### Topic Configuration

- Set appropriate retention periods using `retention.ms`
- Choose cleanup policies (delete or compact) based on data requirements
- Limit the number of topics to manage resource utilization effectively

### Message Structure

- Include timestamps for tracking and ordering
- Use consistent schemas to facilitate processing and evolution
- Incorporate meaningful keys to maintain message grouping and ordering

### Error Handling

- Implement retry mechanisms with exponential backoff strategies
- Validate messages before publishing to prevent processing errors
- Maintain comprehensive logs for monitoring and debugging
- Use Dead Letter Queue (DLQ) for failed messages

### Performance Considerations

- Keep messages under 1MB to optimize throughput
- Utilize compression to minimize message size
- Monitor producer and consumer throughput to identify bottlenecks
- Implement backpressure handling for high-volume scenarios

### Rate Limits

- Maximum message size: 1MB
- Maximum batch size: 1000 messages
- Maximum topics per cluster: 1000
- Maximum consumer groups: 1000
- Maximum retention period: 7 days

## Development

### Project Structure

```
messaging_service/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   └── messaging_service/
│       ├── handler/
│       ├── service/
│       └── repository/
├── pkg/
│   ├── kafka/
│   └── logger/
├── protos/
├── docs/
├── docker/
├── go.mod
├── go.sum
└── docker-compose.yml
```

### Running Tests

```bash
# Run all tests
make test

# Run specific test
go test ./internal/messaging_service/...

# Run with coverage
make test-coverage
```

### Code Generation

```bash
# Generate protobuf files
make proto

# Generate mocks
make mocks
```

## Monitoring

### Metrics

The service exposes Prometheus metrics at `/metrics`:

- Message publish rate
- Message consume rate
- Consumer lag
- Error rates
- Processing latency

### Logging

Logs are available in JSON format with the following levels:
- ERROR: For errors that need immediate attention
- WARN: For potentially harmful situations
- INFO: For general operational information
- DEBUG: For detailed debugging information

## Troubleshooting

### Common Issues

1. **Connection Issues**
   - Check Kafka broker availability
   - Verify network connectivity
   - Check firewall settings

2. **Performance Issues**
   - Monitor consumer lag
   - Check message size
   - Verify batch settings

3. **Message Loss**
   - Check retention settings
   - Verify consumer group settings
   - Monitor DLQ

### Debugging

1. **Enable Debug Logging:**
```bash
export LOG_LEVEL=debug
```

2. **Check Kafka Logs:**
```bash
docker-compose logs kafka
```

3. **Monitor Consumer Groups:**
```bash
kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --all-groups
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support, please:
1. Check the documentation
2. Search existing issues
3. Create a new issue if needed 