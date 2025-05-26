# ---- Stage 1: Builder ----
FROM golang:1.24 AS builder

# Set Go environment
WORKDIR /app

# Set up workspace structure
COPY encryption_service/docker-go.work ./go.work
COPY encryption_service/go.mod encryption_service/go.sum ./encryption_service/
COPY protos/go.mod protos/go.sum ./protos/
COPY encryption_service ./encryption_service
COPY protos ./protos

# Build the encryption service
WORKDIR /app/encryption_service
RUN go mod tidy
RUN go build -o encryption_service ./cmd/server

# ---- Stage 2: Runtime ----
FROM ubuntu:22.04

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy built binary and env
COPY --from=builder /app/encryption_service/encryption_service .
COPY --from=builder /app/.env .

# Expose ports
EXPOSE 8002
EXPOSE 500

RUN chmod +x ./encryption_service

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o encryption-service ./cmd/server

# Use a minimal alpine image for the final container
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/encryption-service .

# Expose the port
EXPOSE 8002
EXPOSE 50502

# Run the application
CMD ["./encryption-service"] 
