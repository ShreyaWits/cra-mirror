# Template Services

A microservice template built with Go, featuring both HTTP and gRPC endpoints, using YugabyteDB for persistence and Redis for caching.

## Features

- HTTP API using Fiber framework
- gRPC server implementation
- YugabyteDB integration for data persistence
- Redis caching layer
- Dependency Injection pattern
- Clean Architecture structure
- Environment-based configuration
- Request validation middleware
- Error handling with custom error codes

## Prerequisites

- Go 1.19 or higher
- YugabyteDB
- Redis
- Protocol Buffers compiler (protoc)

## Project Structure

```
template-services/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── configs/
│   ├── di/
│   ├── pkg/
│   │   ├── cache/
│   │   ├── db/
│   │   └── errors/
│   └── template/
│       ├── handler/
│       ├── middleware/
│       ├── repository/
│       ├── routes/
│       └── service/
├── proto/
└── readme.md
```

## Configuration

The service uses environment variables for configuration. Create a `.env` file with the following variables:

```env
# YugabyteDB Configuration
YUGABYTEDB_HOST=localhost
YUGABYTEDB_PORT=5433
YUGABYTEDB_USER=postgres
YUGABYTEDB_PASSWORD=postgres
YUGABYTEDB_NAME=template_db

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Server Configuration
HTTP_PORT=8080
GRPC_PORT=50051
```

## Running the Service

1. Start the dependencies:
   ```bash
   # Start YugabyteDB
   docker run -d --name yugabyte -p 5433:5433 -p 9042:9042 yugabytedb/yugabyte:latest

   # Start Redis
   docker run -d --name redis -p 6379:6379 redis:latest
   ```

2. Build and run the service:
   ```bash
   go build -o template-service cmd/server/main.go
   ./template-service
   ```

## API Endpoints

### HTTP Endpoints

#### Create Template
```http
POST /v1/templates
Content-Type: application/json

{
    "name": "welcome_email",
    "channel": "email",
    "language": "en",
    "content": "Hello {{name}}!",
    "is_active": true
}
```

#### Get Template
```http
GET /v1/templates?name=welcome_email&channel=email&language=en
```

#### Update Template
```http
PUT /v1/templates/:id
Content-Type: application/json

{
    "is_active": false
}
```

#### Delete Template
```http
DELETE /v1/templates/:id
```

### gRPC Endpoints

#### Get Template V1
```protobuf
rpc GetTemplateV1(GetTemplateRequest) returns (TemplateResponse)

message GetTemplateRequest {
  string name = 1;
  string channel = 2;
  string language = 3;
}
```

### Error Codes

The service uses custom error codes for different scenarios:

- TMP001: Invalid request body
- TMP002: Invalid UUID
- TMP003: Template creation failed
- TMP004: Template fetch failed
- TMP005: Template update failed
- TMP006: Template not found
- TMP007: Template deletion failed
- TMP008: Template ID not found
- TMP009: Missing name
- TMP0010: Missing channel
- TMP0011: Missing language
- TMP0012: Missing content
- TMP0013: Missing is_active status
- TMP0014: Missing template ID
- TMP0015: Template name already exists

## Development

### Generating Protocol Buffers
```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/template.proto
```

### Running Tests
```bash
go test ./...
```

## Versioning

- v1.0.0 - Initial release
  - Basic CRUD operations
  - HTTP and gRPC endpoints
  - YugabyteDB integration
  - Redis caching
  - Request validation
  - Custom error handling

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
