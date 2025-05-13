# CRA (Cloud Resource Automation)

A robust, production-ready cloud resource automation system with Kafka-based messaging for reliable event processing.

## Architecture

### Messaging System
The system uses Apache Kafka with ZooKeeper for reliable message processing:

- **Producer**
  - At-least-once delivery semantics
  - Exactly-once delivery for critical operations
  - Configurable batching and compression
  - Automatic retry with backoff

- **Consumer**
  - Consumer group support
  - Configurable offset management
  - Automatic rebalancing
  - Dead Letter Queue (DLQ) support

- **Topic Management**
  - Automatic topic creation
  - Configurable retention policies
  - Proper partition management
  - Replication factor control

### Key Components

1. **Messaging Service**
   - Kafka-based message processing
   - ZooKeeper for cluster coordination
   - Configurable delivery semantics
   - Consumer group management

2. **Resource Management**
   - Cloud resource provisioning
   - State management
   - Event-driven updates
   - Resource lifecycle handling

3. **Event Processing**
   - Reliable message delivery
   - Event ordering
   - Error handling
   - Retry mechanisms

## Prerequisites

- Go 1.21 or later
- Apache Kafka 2.8 or later
- Apache ZooKeeper 3.7 or later
- Docker and Docker Compose
- Make (optional, for using Makefile commands)
- Protocol Buffers compiler (protoc)

## Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/CRA.git
   cd CRA
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Generate protobuf files**
   ```bash
   make proto
   ```

4. **Start dependencies**
   ```bash
   docker-compose up -d
   ```

5. **Build the project**
   ```bash
   make build
   ```

6. **Run the service**
   ```bash
   ./bin/cra
   ```

## Configuration

### Kafka Configuration
```yaml
kafka:
  brokers:
    - localhost:9092
  zookeeper:
    - localhost:2181
  topics:
    resource_events:
      partitions: 3
      replication_factor: 3
      retention_ms: 604800000  # 7 days
```

### Service Configuration
```yaml
service:
  name: cra
  port: 8080
  log_level: info
  metrics_port: 9090
```

### Environment Variables
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
CRA/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── messaging_service/
│   │   ├── handler/
│   │   ├── service/
│   │   └── repository/
│   └── resource_manager/
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
- Resource operation metrics
- System health metrics

### Logging

Logs are available in JSON format with the following levels:
- ERROR: For errors that need immediate attention
- WARN: For potentially harmful situations
- INFO: For general operational information
- DEBUG: For detailed debugging information

## Security

### Authentication
- SASL/PLAIN
- SASL/SCRAM
- SSL/TLS

### Authorization
- ACLs
- RBAC
- ZooKeeper ACLs

### Encryption
- SSL/TLS
- Secure communication

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

## Deployment

### Docker
```bash
docker build -t cra .
docker run -p 8080:8080 cra
```

### Kubernetes
```bash
kubectl apply -f k8s/
```

## Contributing

1. Fork the repository
2. Create your feature branch
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
