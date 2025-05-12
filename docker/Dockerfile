# ---- Stage 1: Builder ----
    FROM golang:1.24-alpine AS builder

    # Install git and SSL certs for private repos or HTTPS downloads
    RUN apk add --no-cache git ca-certificates
    
    # Set working directory
    WORKDIR /app
    
    # Copy workspace & modules
    COPY go.work go.work.sum ./
    COPY messaging_service/go.mod messaging_service/go.sum ./messaging_service/
    COPY protos/go.mod protos/go.sum ./protos/
    
    # Copy source code
    COPY messaging_service ./messaging_service
    COPY protos ./protos
    COPY messaging_service/.env ./.env
    
    # Download dependencies in workspace mode
    WORKDIR /app/messaging_service
    RUN go mod tidy
    
    # Build binary
    RUN CGO_ENABLED=0 GOOS=linux go build -o messaging_service ./server
    
    # ---- Stage 2: Minimal Image ----
    FROM alpine:latest
    
    WORKDIR /app
    
    # Copy built binary from builder
    COPY --from=builder /app/messaging_service/messaging_service .
    # Copy env file from builder 
    COPY --from=builder /app/.env .
    
    RUN chmod +x ./messaging_service
    
    # Run the binary
    CMD ["./messaging_service"]