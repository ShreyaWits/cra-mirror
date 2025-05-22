# Start from the official Golang base image
FROM golang:1.24-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /build

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Set necessary environment variables needed for our image and build the Go app
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN go build -ldflags="-s -w" -o main ./cmd/server

# Start a new stage from scratch
FROM scratch

# copy the ca-certificate.crt from the build stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /build/main /main
COPY --from=builder /build/.env .env

# Expose port 8005 to the outside world
EXPOSE 8005
EXPOSE 50505

# Command to run the executable
ENTRYPOINT ["/main"]