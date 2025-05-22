# Start from the official Golang bullseye base image
FROM golang:1.23-alpine AS builder



# Set the current working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files from the current directory (apps/v1)
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code from the current directory
COPY . .

# Set environment variables for Go build
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Build the Go binary with optimization flags to reduce size
RUN go build -ldflags="-s -w" -o main ./cmd/worker

# Final stage - build a minimal Docker image with the compiled binary
FROM scratch

# Copy CA certificates and compiled binary from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/main /main
COPY --from=builder /build/.env .env


# Command to run the executable
ENTRYPOINT ["/main"]
