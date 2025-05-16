# Encryption Microservice

A high-performance Go microservice for encryption and decryption operations using AES-256-GCM. This service provides a secure way to encrypt sensitive data using envelope encryption, where data is encrypted with a Data Encryption Key (DEK) that is itself encrypted with a Key Encryption Key (KEK).

## Features

- AES-256-GCM encryption/decryption for secure data handling
- Key Encryption Key (KEK) management through external KMS
- Data Encryption Key (DEK) generation with secure random values
- Encrypted DEK (EDEK) generation for secure key storage
- REST API endpoints for encryption operations
- Stateless design for high scalability
- Secure key handling with memory cleanup

## Project Structure

```
project-name/
├── Dockerfile                # Main Dockerfile for the application
├── README.md                 # Project documentation
├── docker-compose.yml        # Docker Compose configuration
├── go.mod                    # Go module file
├── go.sum                    # Go module checksums
├── README.md            # Directory documentation
├── cmd/                      # Application entry points
│   └── server/              # Server application entrypoint
│       └── main.go          # Main server entry point
├── internal/                # Private application code
│   ├── app/                # Application core
│   │   └── router.go       # API router configuration
│   ├── common/             # Common utilities and shared code
│   ├── configs/            # Configuration handlers
│   └── modules/            # Feature modules
│       └── [module-name]/  # Individual feature modules
│           ├── apis/       # HTTP handlers and routes
│           ├── constants/  # Module-specific constants
│           ├── di/        # Dependency injection setup
│           ├── models/    # Data models/entities
│           ├── repositories/ # Data access layer
│           ├── services/  # Business logic
│           └── utils/     # Module-specific utilities
├── pkg/                  # Shared, exportable packages
│   ├── crypto/            # crypto encryprion client library
│   ├── kms/               # kms client library
└── scripts/              # Utility scripts
```

## API Endpoints

### Health Check
- `GET /health` - Check service health
  ```json
  Response:
  {
    "status": "healthy",
    "version": "1.0.0"
  }
  ```

### Encryption Operations
--

- `GRPC /HealthCheck` - Check Service Health
  ```json

  Response:
  {
    "status":"OK"
  }
  ```

- `GRPC /Encrypt` - Encrypt data
  ```json
  Request:
  {
   "token":"test token",
   "data": [
      {"fields" :
         {"passwords":"xxxxx", "e_type":"public"},
      },
      {"fields" :
         {"accountNumber":"xxxxx", "e_type":"private"},
      },
   
   ]
  }
  
  Response:
  {
    "data": [
       {"fields" :
         {"passwords":"public_xxyxy"},
       },
          {"fields" :{
         {"accountNumber":"private_xxyxy"},
          }
       }
      ],
  }
  ```

- `GRPC /Decrypt` - Decrypt data
  ```json
  Request:
   {
   "token":"test token",
   "userId":"userId",
   "data": [
         {"fields" :
             {"passwords":"public_xxyxy"},
         },
         {"fields" :
            {"accountNumber":"private_xxyxy"},
         }
        
      ]
  }

  Response:
    {
    "data": [
         {"fields" :
      {"passwords":"xxxxx"},
         },
          {"fields" :
            {"accountNumber":"xxxxx"},}
      ]
  }
  ```

- `GRPC /GenerateEDEK` - Generate Encrypted DEK
  ```json

  Request:
  {
   "token":"test token"
   }

  Response:
  {
    "edekPrivate": "base64 encoded encrypted data key",
    "edekPublic": "base64 encoded encrypted data key"
  }
  ```

## Configuration

The service can be configured using environment variables:

```env
# Server Configuration
PORT=50051
ENV=development
```

## Getting Started

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Set up environment variables:
   ```bash
   cp .env.example .env
   ```
4. Run the service:
   ```bash
   go run cmd/server/main.go
   ```

## Docker Deployment

The service can be deployed using Docker and Docker Compose, which includes both the encryption service and a HashiCorp Vault instance for key management.

### Using Docker Compose

1. Build and start the services:
   ```bash
   docker-compose up --build
   ```

2. Stop the services:
   ```bash
   docker-compose down
   ```

### Services

The Docker Compose setup includes:

1. **Encryption Service**
   - Port: 8082 (mapped to container port 8080)
   - Environment variables:
     - `VAULT_ADDR`: Vault server address
     - `VAULT_TOKEN`: Vault authentication token
     - `VAULT_PATH`: Vault secrets engine path

2. **HashiCorp Vault**
   - Port: 8200
   - Development mode with root token
   - Transit secrets engine enabled
   - Persistent volume for data storage
   - Health checks configured

### Docker Image

The service uses a multi-stage build process:
1. Builder stage:
   - Uses `golang:1.22-alpine` as base
   - Compiles the application
2. Final stage:
   - Uses minimal `alpine` image
   - Contains only the compiled binary
   - Exposes port 8080

### Environment Variables

When running in Docker, the following environment variables are required:
```env
VAULT_ADDR=http://vault:8200
VAULT_TOKEN=root
VAULT_PATH=transit
```

## Security Considerations

- All encryption operations use AES-256-GCM for authenticated encryption
- Keys are managed through a Key Management System (KMS) for secure key storage
- Input validation and sanitization to prevent injection attacks
- Rate limiting to prevent abuse
- HTTPS enforcement for all API endpoints
- Memory cleanup after key operations
- No persistent storage of encryption keys
- Secure random number generation for key material

## Implementation Details

### Encryption Flow
1. Generate a random DEK using secure random number generation
2. Retrieve KEK from KMS
3. Encrypt the DEK with KEK to create EDEK
4. Use DEK to encrypt the data using AES-256-GCM
5. Return encrypted data and EDEK
6. Clear DEK from memory

### Decryption Flow
1. Retrieve KEK from KMS
2. Decrypt EDEK to obtain DEK
3. Use DEK to decrypt the data
4. Clear DEK from memory
5. Return decrypted data

### Generation Flow
1. **private (End-to-End) Key Generation**
   - Generate a random DEK_private using secure random number generation
   - Retrieve KEK_private from KMS
   - Encrypt DEK_private with KEK_private to create EDEK_private
   - Store EDEK_private in user service
   - Clear DEK_private from memory

2. **public Key Generation**
   - Generate a random DEK_public using secure random number generation
   - Retrieve KEK_public from KMS
   - Encrypt DEK_public with KEK_public to create EDEK_public
   - Store EDEK_public in user service
   - Clear DEK_public from memory

3. **Key Usage**
   - private keys are used for sensitive data like passwords
   - public keys are used for data that needs to be accessible by multiple parties
   - Each data field is prefixed with its type (private_ or public_) for identification

4. **Key Generation**
   - `GenerateKEK`: Creates a new Key Encryption Key (KEK)
   - Supports different key types based on role (private or public)


### Key Management System (KMS) Implementation

The KMS package provides a flexible interface for key management operations. It supports different KMS implementations through a common interface:

```go
type KmsService interface {
    StoreKEK(ctx context.Context, kekID string, kek []byte, role string) error
    RetrieveKEK(ctx context.Context, kekID string, role string) ([]byte, error)
}
```

#### KMS Operations


1. **Key Storage**
   - `StoreKEK`: Securely stores KEKs with role-based access control
   - Keys are stored with associated metadata and role information

2. **Key Retrieval**
   - `RetrieveKEK`: Retrieves KEKs based on ID and role
   - Implements role-based access control for key retrieval



#### Role-Based Access Control

The KMS implementation supports two main roles:
- `private`: For end-to-end encryption keys
- `public`: For public encryption keys

Each operation is role-aware and ensures proper access control based on the key's role.

## License

MIT License

## Support

For support and questions, please contact the development team.

