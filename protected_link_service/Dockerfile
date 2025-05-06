# ---------- Builder Stage ----------
    FROM golang:1.24-alpine AS builder

    # Set working directory
    WORKDIR /app
    
    # Copy go mod and sum files first for dependency caching
    COPY go.mod go.sum ./
    
    # Download dependencies
    RUN go mod download
    
    # Copy the entire source code
    COPY . .
    
    # Build the application binary
    RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o protected_link ./cmd/server/main.go
    
    # ---------- Final Stage ----------
    FROM alpine:latest
    
    WORKDIR /app
    
    # Add CA certificates (required for HTTPS)
    RUN apk --no-cache add ca-certificates
    
    # Copy the compiled binary from builder
    COPY --from=builder /app/protected_link .
    
    # ✅ Copy the internal config files needed at runtime
    COPY --from=builder /app/internal/configs /app/internal/configs

    # Give execute permission (optional if already executable)
    RUN chmod +x /app/protected_link
    
    # Expose application port
    EXPOSE 8080
    
    # Start the application
    CMD ["/app/protected_link"]
    