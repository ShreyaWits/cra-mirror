# ---- Stage 1: Builder ----
    FROM golang:1.24 AS builder

    # Install librdkafka dependencies
    RUN apt-get update && apt-get install -y \
        build-essential \
        wget \
        git \
        pkg-config \
        curl \
        gnupg \
        ca-certificates
    
    # Install Confluent librdkafka (latest stable via APT)
    RUN curl -fsSL https://packages.confluent.io/deb/7.6/archive.key | gpg --dearmor -o /etc/apt/trusted.gpg.d/confluent.gpg && \
        echo "deb [arch=amd64] https://packages.confluent.io/deb/7.6 stable main" > /etc/apt/sources.list.d/confluent.list && \
        apt-get update && apt-get install -y librdkafka-dev
    
    # Set Go environment
    WORKDIR /app
    COPY go.work go.work.sum ./
    COPY messaging_service/go.mod messaging_service/go.sum ./messaging_service/
    COPY protos/go.mod protos/go.sum ./protos/
    COPY messaging_service ./messaging_service
    COPY protos ./protos
    COPY messaging_service/.env ./.env
    
    WORKDIR /app/messaging_service
    RUN go mod tidy
    RUN go build -o messaging_service ./cmd/server
    
    # ---- Stage 2: Runtime ----
    FROM ubuntu:22.04
    
    # Install runtime dependencies
    RUN apt-get update && apt-get install -y \
        ca-certificates \
        librdkafka1 \
        && rm -rf /var/lib/apt/lists/*
    
    WORKDIR /app
    
    # Copy built binary and env
    COPY --from=builder /app/messaging_service/messaging_service .
    COPY --from=builder /app/.env .
    
    RUN chmod +x ./messaging_service
    
    CMD ["./messaging_service"]