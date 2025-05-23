# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copy Go workspace and module files
COPY document_processing/docker-go.work ./go.work
COPY document_processing/go.mod document_processing/go.sum ./document_processing/
COPY protos/go.mod protos/go.sum ./protos/
COPY document_processing ./document_processing
COPY protos ./protos

WORKDIR /app/document_processing

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/document_processing/document_processing ./cmd/server/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /app/document_processing/document_processing .

# Copy any additional required files (otel config, but NOT .env)
COPY --from=builder /app/document_processing/otel-config.yaml .

EXPOSE 50508
EXPOSE 8008

CMD ["./document_processing"]