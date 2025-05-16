# NPS Notification Service

A robust notification service built with Go, designed to handle various types of notifications including email, SMS, push notifications, and more.

## 🚀 Features

Multiple notification channels support:

- Email (SendGrid)
- SMS (Twilio)
- Push Notifications (Firebase Cloud Messaging)
- In-app notifications
  Message queuing with Kafka
  Caching with Redis
  Database storage with Cassandra
  Distributed tracing with OpenTelemetry
  Workflow orchestration with Temporal
  RESTful API with Fiber framework

## 🛠️ Tech Stack

**Language**: Go 1.24.1
**Web Framework**: Fiber v2
**Database**: Cassandra
**Cache**: Redis
**Message Queue**: Kafka
**Email Service**: SendGrid
**SMS Service**: Twilio
**Push Notifications**: Firebase Cloud Messaging
**Workflow Engine**: Temporal
**Monitoring**: OpenTelemetry with Jaeger
**Containerization**: Docker

## 📦 Dependencies

### Core Dependencies

`github.com/gofiber/fiber/v2` - Web framework
`github.com/gocql/gocql` - Cassandra client
`github.com/redis/go-redis/v9` - Redis client
`github.com/segmentio/kafka-go` - Kafka client
`github.com/sendgrid/sendgrid-go` - Email service
`github.com/twilio/twilio-go` - SMS service
`firebase.google.com/go/v4` - Firebase Cloud Messaging
`go.temporal.io/sdk` - Temporal workflow engine
`go.opentelemetry.io/otel` - OpenTelemetry for tracing

### Development Dependencies

`github.com/stretchr/testify` - Testing framework
`github.com/joho/godotenv` - Environment configuration
`github.com/sirupsen/logrus` - Logging

## 🏗️ Project Structure

notification-service/
├── cmd/ # Application entry points
├── internal/ # Private application code
│ ├── app/ # Application setup and configuration
│ ├── configs/ # Configuration management
│ ├── constant/ # Constants and enums
│ ├── domain/ # Domain models and business logic
│ ├── dtos/ # Data Transfer Objects
│ └── service/ # Service layer implementations
├── pkg/ # Public library code
├── proto/ # Protocol buffer definitions
├── cassandra-init/ # Database initialization scripts
├── Dockerfile # Main service Dockerfile
├── Dockerfile.worker # Worker service Dockerfile
├── docker-compose.yml # Production Docker Compose
└── docker-compose.dev.yml # Development Docker Compose

## 🚀 Getting Started

### Prerequisites

Go 1.24.1 or higher
Docker and Docker Compose
Cassandra
Redis
Kafka
Temporal server

### Environment Setup

1. Clone the repository
2. Copy .env.example to .env and configure your environment variables
3. Install dependencies:

bash
go mod download

### Running the Service

#### Using Docker Compose

bash

# Development environment

docker-compose -f docker-compose.dev.yml up

# Production environment

docker-compose up

#### Running Locally

bash

# Start the main service

go run cmd/main.go

# Start the worker service

go run cmd/worker/main.go

## 📝 API Documentation

### gRPC Service

The service exposes a gRPC API for notification configuration and management.

#### Service Definition

protobuf
service NotificationService {
rpc Configure(ConfigureRequest) returns (ConfigureResponse);
rpc GetConfigure(ConfigureRequest) returns (GetConfigureResponse);
}

#### Message Types

##### ChannelConfig

protobuf
message ChannelConfig {
string service = 1; // e.g., email, whatsapp, sms
string primary = 2; // e.g., sendgrid
string fallback = 3; // e.g., smtp
}

##### ConfigureRequest

protobuf
message ConfigureRequest {
repeated ChannelConfig configs = 1; // List of channel configurations
}

##### ConfigureResponse

protobuf
message ConfigureResponse {
string status = 1;
int32 statusCode = 2;
string message = 3;
FieldError fieldError = 4;
string data = 5;
}

##### GetConfigureResponse

protobuf
message GetConfigureResponse {
repeated ChannelConfig configs = 1;
string status = 2;
int32 statusCode = 3;
string message = 4;
FieldError fieldError = 5;
string data = 6;
}

##### FieldError

protobuf
message FieldError {
string field = 1; // The field where the error occurred
string errorMsg = 2; // The error message
string errorCode = 3; // The custom error code
}

### HTTP Endpoints

#### Health Check

**Endpoint**: /health
**Method**: GET
**Response**: Service health status

#### Notification Configuration

**Endpoint**: /api/v1/notifications/configure
**Method**: POST
**Request Body**:

json
{
"configs": [
{
"service": "email",
"primary": "sendgrid",
"fallback": "smtp"
},
{
"service": "sms",
"primary": "twilio",
"fallback": "firebase"
}
]
}

**Response**:

json
{
"status": "success",
"statusCode": 200,
"message": "Configuration updated successfully",
"data": null
}

#### Get Notification Configuration

**Endpoint**: /api/v1/notifications/configure
**Method**: GET
**Response**:

json
{
"configs": [
{
"service": "email",
"primary": "sendgrid",
"fallback": "smtp"
},
{
"service": "sms",
"primary": "twilio",
"fallback": "firebase"
}
],
"status": "success",
"statusCode": 200,
"message": "Configuration retrieved successfully",
"data": null
}

### Error Responses

All endpoints may return the following error structure:
json
{
"status": "error",
"statusCode": 400,
"message": "Invalid request",
"fieldError": {
"field": "service",
"errorMsg": "Service type is required",
"errorCode": "REQUIRED_FIELD"
}
}

Common HTTP Status Codes:
200: Success
400: Bad Request
401: Unauthorized
403: Forbidden
404: Not Found
500: Internal Server Error

## 🔧 Configuration

The service can be configured using environment variables or a configuration file. Key configuration parameters include:

Database connection settings
Redis connection details
Kafka broker configuration
Service credentials (SendGrid, Twilio, Firebase)
Temporal workflow settings
OpenTelemetry configuration

## 📊 Monitoring

The service uses OpenTelemetry for distributed tracing and monitoring. Metrics and traces can be viewed in Jaeger.
