# Document Processing Service

This is a Go-based microservice for processing documents, including tasks like file storage, data extraction using AI models (Gemini, Llama), and storing processed data.

## Services

The service interacts with the following dependencies:

- **MinIO:** Used for storing uploaded document files. (Service name in Docker Compose: `minio`)
- **YugabyteDB:** Used for storing metadata or extracted data related to document processing. (Assumed external container name: `yugabyte-db`)
- **Config Service:** An external service providing configuration for the Document Processing service. (Assumed external container name: `nps-config-service`)

## Prerequisites

- Docker and Docker Compose installed.
- Access to the external `config-service` and its associated YugabyteDB container, running on a Docker network accessible to this service (configured to use network `config_service_nps2.0`).

## Getting Started

1.  **Navigate to the service directory:**

    ```bash
    cd Document-Processing
    ```

2.  **Ensure external services are running:**
    Make sure your `config-service` container (`nps-config-service`) and its associated YugabyteDB container (`yugabyte-db`) are already running and connected to the `config_service_nps2.0` Docker network.

3.  **Configure Environment Variables:**
    Ensure the necessary environment variables are set for the `document-processing` service, especially for connecting to YugabyteDB and MinIO. You can do this by creating a `.env` file in the `Document-Processing` directory or by exporting them in your shell.

    Example `.env` content:

    ```env
    # Database Configuration (YugabyteDB)
    YUGABYTE_HOST=yugabyte-db
    YUGABYTE_PORT=5433
    YUGABYTE_USER=your_db_user
    YUGABYTE_PASSWORD=your_db_password
    YUGABYTE_NAME=your_db_name

    # MinIO Configuration
    MINIO_ENDPOINT=minio:9000
    MINIO_ACCESS_KEY=minioadmin
    MINIO_SECRET_KEY=minioadmin
    MINIO_BUCKET_NAME=document-processing
    MINIO_USE_SSL=false # Set to true if using SSL

    # Gemini API Key
    GEMINI_API_KEY=your_gemini_api_key

    # Server Port (internal to container)
    SERVER_PORT=50051

    # Configuration Service URL (if used - currently commented out in docker-compose)
    # CONFIG_SERVICE_URL=http://nps-config-service:4001/api/v1/config/dev/document-processing

    # Environment (used by config loading if fetching from API)
    # ENVIRONMENT=dev
    ```
    Replace placeholder values with your actual credentials and configuration.

4.  **Build and Run with Docker Compose:**
    In the `Document-Processing` directory, run the following command to build the document-processing and minio images and start the containers:

    ```bash
    docker-compose up --build
    ```

## Project Structure (Key Directories)

- `cmd/server/`: Contains the main entry point for the gRPC server (`main.go`).
- `internal/config/`: Handles loading configuration from environment variables and potentially an external API.
- `internal/repository/`: Contains data access logic, including the MinIO repository.
- `internal/services/`: Contains the core business logic and AI model interactions (Gemini, Llama).
- `proto/`: Contains the gRPC service definitions (`.proto` files).

## Accessing the Service

The gRPC service will be accessible on your host machine at `localhost:8080`.

## Testing with Postman

1.  Open Postman and create a new gRPC request.
2.  Set the server address to `localhost:8080`.
3.  Import the `.proto` file(s) from the `proto/` directory.
4.  Select the desired service and method.
5.  Provide the request payload and send the request.

## Troubleshooting

- If you encounter a "port is already allocated" error, ensure no other process or container is using the required ports (e.g., 5433, 8080, 9000, 9001).
- If your service fails to connect to YugabyteDB or MinIO, double-check the environment variables in your `.env` file and the container names/network configuration in `docker-compose.yml`.
- For "call canceled" errors in Postman, check the logs of the `document-processing` container (`docker logs document-processing-document-processing-1`) for server-side errors.
