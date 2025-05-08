FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download


COPY .env ./

# Copy source code
COPY . .


# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o messaging_service ./server

# Use a minimal alpine image for the final container
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/messaging_service .

# Ensure the binary is executable
RUN chmod +x ./messaging_service

# Verify the binary is present and executable
RUN ls -l ./messaging_service

# Run the application
CMD ["./messaging_service"] 