FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git
RUN apk add --no-cache git

# Set GOPROXY to use a more reliable proxy
ENV GOPROXY=https://goproxy.io,direct

COPY . .
COPY go.mod go.sum ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/.env .env
EXPOSE 50051 8080
CMD ["./server"] 