# --- Builder Stage ---
    FROM golang:1.24 AS builder

    # Set the working directory inside the container
    WORKDIR /app
    
    # Copy go mod and sum files
    COPY go.mod go.sum ./
    
    # Download all dependencies
    RUN go mod download
    
    # Copy the source code from the current directory to the working directory
    COPY . ./
    
    # Build the application
    RUN go build -o main cmd/server/main.go
    
    # --- Final Stage ---
    FROM ubuntu:latest
    
    # Set the working directory inside the container
    WORKDIR /app
    
    # Copy the binary from the builder stage
    COPY --from=builder /app/main /app/main
    COPY .env /app/.env
    RUN ls -la /app
    
    # Expose the port the app listens on
    EXPOSE 50051
    
    # Command to run when the container starts
    CMD ["/app/main"]
    