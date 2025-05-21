# Protected Link Service

A microservice for generating and managing protected links with authentication and access control features.

## Overview

Protected Link Service is a Go-based microservice that provides secure link generation with customizable access controls. It uses Cassandra for persistent storage, Redis for caching, and Kafka for event streaming.

## Features

- Generate protected links with customizable security settings
- Support for OTP authentication
- Configurable expiration times
- Multiple authentication channels (email, phone, etc.)
- gRPC interfaces
- Support for different request types (auth, payment, file)
- Hybrid and JWT model types for data storage

## Technology Stack

- **Language**: Go
- **Database**: Apache Cassandra
- **Cache**: Redis
- **Messaging**: Apache Kafka
- **API**: gRPC

## Prerequisites

- Docker and Docker Compose
- Go 1.19+ (for local development)

## Running Locally with Docker

1. Clone the repository
2. Navigate to the project directory
3. Run using Docker Compose:

```bash
docker compose -f docker-compose.local.yml up -d
```

## Environment Variables

The service can be configured using the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | HTTP server port | 8080 |
| GRPC_PORT | gRPC server port | 50051 |
| DB_PASSWORD | Redis and database password | password |
| CASSANDRA_HOST | Cassandra host | cassandra |
| CASSANDRA_PORT | Cassandra port | 9042 |
| CASSANDRA_KEYSPACE | Cassandra keyspace | protectedlink |
| REDIS_HOST | Redis host | redis |
| REDIS_PORT | Redis port | 6379 |
| KAFKA_HOST | Kafka host | kafka |
| KAFKA_PORT | Kafka port | 9092 |
| APP_ENV | Application environment | development |

## API Documentation

### gRPC Services

The service exposes the following gRPC methods:

#### 1. GenerateUrlV1

Generates a secure URL based on the request.

**Request Payload:**
```json
{
  "user_id": "76677688",
  "request_type": "auth|payment|file",
  "model_type": "hybrid|jwt",
  "phone_number": "+91 8764346732",
  "expire_in": "24h",
  "email": "test@gmail.com",
  "channel_type": "email",
  "otp_required": true,
  "data": { ... } // Varies by type
}
```

**Data Examples by Request Type:**

Auth:
```json
"data": {
  "auth_provider": "Test data",
  "permissions": [
    "read",
    "write"
  ]
}
```

Payment:
```json
"data": {
  "transaction_id": "TXN123456",
  "amount": 999.99,
  "currency": "INR"
}
```

File:
```json
"data": {
  "file_url": "https://example.com/file.pdf",
  "file_type": "pdf",
  "size": "1MB"
}
```

**Response:**
```json
{
  "message": "URL generated successfully",
  "success": true,
  "code": 200,
  "data": {
    "redirectUrl": "http://localhost:8081?token=eyJhbGciO"
  }
}
```

#### 2. GetUrlDataV1

Retrieves and decrypts secure data based on the URL token.

**Request Payload:**
```json
{
  "token": "token"
}
```

**Response (OTP not required):**
```json
{
  "message": "URL data retrieved successfully",
  "success": true,
  "code": 200,
  "data": {
    "request_type": "auth",
    "data": {
      "auth_provider": "Test data",
      "permissions": [
        "read",
        "write"
      ],
      "user_id": "76677688"
    }
  }
}
```

**Response (OTP required):**
```json
{
  "message": "OTP sent successfully",
  "success": true,
  "code": 202,
  "data": {
    "phone_number": "+919999999999",
    "expires_in": "24h"
  }
}
```

#### 3. VerifyOtpV1

Authenticates user with OTP and returns data.

**Request Payload:**
```json
{
  "user_id": "76677688",
  "otp": "516550",
  "verification_id": "14518c26-840c-4c73-9e3d-80fe19ba0b78-76677688"
}
```

**Success Response:**
```json
{
  "message": "URL data retrieved successfully",
  "success": true,
  "code": 200,
  "data": {
    "request_type": "auth",
    "data": {
      "auth_provider": "Test Data",
      "permissions": [
        "read",
        "write"
      ],
      "user_id": "76677688"
    }
  }
}
```

**Failure Response:**
```json
{
  "message": "Invalid or expired OTP",
  "success": false,
  "code": 401
}
```

#### 4. RevokeUrlV1

Revokes a previously generated URL.

**Request Payload:**
```json
{
  "link": "eyJhbGciO"
}
```

**Success Response:**
```json
{
  "message": "URL revoked successfully",
  "success": true,
  "code": 200
}
```

**Failure Responses:**

Missing token:
```json
{
  "message": "Either token must be provided",
  "success": false,
  "code": 400
}
```

Invalid token:
```json
{
  "message": "Invalid or expired token",
  "success": false,
  "code": 401
}
```

URL already revoked:
```json
{
  "message": "No matching URL found to revoke",
  "success": false,
  "code": 404
}
```

## Service Architecture

### Data Flow

1. **URL Generation**:
   - Validate user token
   - Prepare API data based on request_type
   - Encrypt data
   - Store in Cassandra if model_type is hybrid
   - Generate secure token
   - Return redirect URL

2. **URL Data Retrieval**:
   - Validate token
   - If OTP required, trigger OTP delivery
   - Retrieve encrypted data (from token or Cassandra)
   - Decrypt data and return

3. **OTP Verification**:
   - Validate OTP for given reference key
   - If valid, retrieve and decrypt data
   - Return data in response

4. **URL Revocation**:
   - Validate token
   - Clear token from Redis
   - Remove data from Cassandra if in hybrid mode
   - Return confirmation

### Required Interfaces

- **AllocateMemory()** from Memory Manager: For storing tokens, session details, or temporary OTP data
- **SendOTP()** from Email Service: For sending OTPs if required during URL generation
- **ProcessVerification()** from Authentication Service: Verifies user authentication before URL generation

## Database Schema

### Cassandra Tables

- `protectedLink` - Stores link information and access settings
- `schema_migrations` - Tracks database migrations

## Development

### Project Structure

```
protected_link_service/
├── cmd/                # Application entry points
├── internal/           # Internal packages
│   ├── configs/        # Configuration handling
│   └── modules/        # Business logic modules
├── pkg/                # Reusable packages
│   ├── cassandra/      # Cassandra client
│   └── redis/          # Redis client
├── migrations/         # Database migrations
└── docker-compose.local.yml  # Local development setup
```

### Running Tests

```bash
go test ./...
```

## Troubleshooting

- **Cassandra Connection Issues**: Ensure the keyspace exists and is properly configured
- **Redis Connection Issues**: Check that Redis host and port are correctly set
- **Kafka Connection Issues**: Verify Kafka broker is running and accessible

## License

[MIT License](LICENSE)