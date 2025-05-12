# Kafka Messaging Service

A high-performance Kafka messaging service implementation with support for topic creation, message publishing, and message subscription.

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

1. **Clone the repository:**
```bash
git clone <repository-url>
cd messaging_service
```

2. **Install dependencies:**
```bash
go mod download
```

3. **Generate protobuf files:**
```bash
make proto
```

4. **Build the service:**
```bash
make build
```

## Configuration

The service can be configured using environment variables or a `.env` file:

```env
# Kafka Configuration
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=my-consumer-group
KAFKA_CLIENT_ID=messaging-service

# Service Configuration
SERVICE_PORT=8080
SERVICE_HOST=localhost
LOG_LEVEL=info

# Performance Tuning
MAX_BATCH_SIZE=1000
MAX_MESSAGE_SIZE=1048576
CONSUMER_POOL_SIZE=10
```

## Running the Service

### Using Docker Compose

1. **Start the service with Kafka:**
```bash
docker-compose up -d
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

## API Usage

### 1. Create a Topic

```bash
curl -X POST http://localhost:8080/v1/topics \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "my-temporary-topic",
    "num_partitions": 1,
    "replication_factor": 1,
    "config": {
      "retention.ms": "3000",
      "cleanup.policy": "delete"
    }
  }'
```

### 2. Publish a Message

```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "my-temporary-topic",
    "value": {
      "action": "login123",
      "timestamp": "2023-10-01T12:00:00Z"
    },
    "headers": {
      "Content-Type": "application/json"
    },
    "key": "user123"
  }'
```

### 3. Subscribe to Messages

```bash
# Using gRPCurl
grpcurl -plaintext -d '{
  "topic": "my-temporary-topic",
  "group_id": "my-consumer-group",
  "consumer_config": {
    "auto_offset_reset": "latest",
    "isolation_level": "read_committed"
  }
}' localhost:8080 messaging.MessagingService/Subscribe
```

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

## API Documentation

For detailed API documentation, see [API.md](docs/API.md). 