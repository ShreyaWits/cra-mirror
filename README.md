# CRA
CRA 2.0

## Protected Link Service

### Docker Setup

The service can be launched using Docker Compose with environment variables.

1. Create a `.env` file in the root directory with the following variables:
```
# Server Configuration
PORT=8080
GRPC_PORT=50051

# Database Configurations
DB_PASSWORD=your_secure_password

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379

# Kafka Configuration
KAFKA_HOST=kafka
KAFKA_PORT=9092

# Zookeeper Configuration
ZOOKEEPER_PORT=2181

# Cassandra Configuration
CASSANDRA_HOST=cassandra
CASSANDRA_PORT=9042
CASSANDRA_CLUSTER_NAME=ProtectedLinkCluster
CASSANDRA_DC=dc1
CASSANDRA_ENDPOINT_SNITCH=GossipingPropertyFileSnitch

# Application Configuration
APP_ENV=development
LOG_LEVEL=info
```

2. Launch the services:
```bash
cd protected_link_service
docker-compose up -d
```

3. To stop the services:
```bash
docker-compose down
```

For development or testing, you can run specific services:
```bash
docker-compose up app redis cassandra
```

### Note on Environmental Configuration

The Docker Compose file is configured to use environment variables with fallback default values. You can either:
- Create a `.env` file (recommended for local development)
- Set environment variables directly in your shell
- The service will use default values if no variables are provided
