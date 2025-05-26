# Start from the official Golang base image
FROM golang:1.24 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the thirdparty_service directory and protos directory from the build context (workspace root)
COPY thirdparty_service /app/thirdparty_service
COPY protos /app/protos

# Change working directory to the service directory for build steps
WORKDIR /app/thirdparty_service

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
# Invalidate cache
RUN go mod download

# Set necessary environment variables needed for our image and build the Go app
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Build the Go app
RUN go build -ldflags="-s -w -extldflags '-static'" -o main ./cmd/server

# Start a new stage from scratch
FROM scratch

# copy the ca-certificate.crt from the build stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/thirdparty_service/main /main
COPY --from=builder /app/thirdparty_service/.env .env

EXPOSE 8080
EXPOSE 9090

# Command to run the executable
ENV IS_DOCKER=false
WORKDIR /
ENTRYPOINT ["/main"]
