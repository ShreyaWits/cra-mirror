# Messaging Service

A scalable, high-performance messaging service built with Go, gRPC, and Kafka.

## Features

- gRPC-based messaging interface
- Kafka topic management API
- Configurable delivery semantics (at-least-once, exactly-once)
- Publisher-subscriber messaging patterns
- Topic creation with custom configurations
- Containerized deployment with Docker

## Architecture

The messaging service is built using:
- **Go** - Core programming language
- **gRPC** - Communication protocol
- **Kafka** - Message broker
- **Docker** - Containerization

## Setup and Installation

### Prerequisites

- Docker and Docker Compose
- Go 1.24+ (for development)
- Protobuf compiler (for development)

### Environment Configuration

Configure your Kafka broker address and other settings in the `.env` file:

```bash
# Kafka configuration
KAFKA_BROKER=kafka:9092
KAFKA_AUTO_CREATE_TOPICS_ENABLE=true

# Service configuration
GRPC_PORT=50051
```

You can also use separate `.env` files for different environments (development, testing, production).

### Running the Service

Clone the repository and start the service using Docker Compose:

```bash
git clone https://github.com/yourusername/cra.git
cd cra
docker compose -f docker/docker-compose.yml up -d
```

The messaging service will be accessible at `localhost:50051`.

## API Reference

### PublishMessage

Publishes a message to a specified Kafka topic.

```proto
rpc PublishMessage(PublishRequest) returns (PublishResponse);
```

**Example**:
```go
client.PublishMessage(ctx, &pb.PublishRequest{
    Topic: "notifications",
    Value: map[string]string{
        "message": "Hello World",
        "priority": "high",
    },
})
```

### SubscribeStream

Streams messages from a Kafka topic to the consumer.

```proto
rpc SubscribeStream(SubscribeRequest) returns (stream KafkaMessage);
```

**Example**:
```go
stream, err := client.SubscribeStream(ctx, &pb.SubscribeRequest{
    Topic: "notifications",
    GroupId: "notification-consumers",
})
for {
    msg, err := stream.Recv()
    if err != nil {
        break
    }
    // Process message
    fmt.Println("Received:", msg.Value["message"])
}
```

### CreateTopic

Creates a new Kafka topic with specified configuration.

```proto
rpc CreateTopic(CreateTopicRequest) returns (CreateTopicResponse);
```

**Example**:
```go
client.CreateTopic(ctx, &pb.CreateTopicRequest{
    Topic: "notifications",
    NumPartitions: 3,
    ReplicationFactor: 1,
    Config: map[string]string{
        "retention.ms": "604800000",          // 7 days retention
        "compression.type": "lz4",
        "cleanup.policy": "delete",
    },
})
```

## Common Topic Configurations

```go
// High-throughput topic with short retention
highThroughputConfig := map[string]string{
    "retention.ms": "3600000",           // 1 hour retention
    "compression.type": "lz4",           // Efficient compression
    "segment.bytes": "536870912"         // 512MB segments
}

// Compacted topic for key-value storage
compactedTopicConfig := map[string]string{
    "cleanup.policy": "compact",         // Keep latest value per key
    "delete.retention.ms": "86400000",   // 24 hours tombstone retention
    "min.compaction.lag.ms": "10000"     // 10 second lag before compaction
}

// Durable topic with longer retention
durableTopicConfig := map[string]string{
    "retention.ms": "2592000000",        // 30 days retention
    "min.insync.replicas": "2",          // Require at least 2 replicas
    "unclean.leader.election.enable": "false"
}
```

## Environment Variables

The service can be configured using the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| GRPC_PORT | gRPC server port | 50051 |
| KAFKA_BROKER | Kafka broker address | kafka:9092 |
| KAFKA_AUTO_CREATE_TOPICS_ENABLE | Enable auto topic creation | true |

## Development

### Building the Service

```bash
cd messaging_service
go mod tidy
go build -o messaging_service ./server
```

### Regenerating Protobuf Code

```bash
cd protos
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative messaging_service/messaging_service.proto
```

## License

MIT
