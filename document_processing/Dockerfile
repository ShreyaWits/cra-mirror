# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o document-processing ./cmd/server/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /app/document-processing .

# Copy any additional required files
COPY --from=builder /app/.env .
COPY --from=builder /app/otel-config.yaml .

# Expose the service port
EXPOSE 50508
EXPOSE 8008

# Run the application
CMD ["./document-processing"]
