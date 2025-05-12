---
# 🧠 Centralized Caching Service

A scalable Redis-backed caching layer for microservices, supporting gRPC-based key-value operations with namespace isolation and TTL-based expiry.
---

## 🚀 Getting Started

### Prerequisites

- Redis (Cluster or standalone)
- Go (for development and running locally)
- Docker + Kubernetes (optional, for deployment)
- `protoc` + Go plugins:

  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```

---

### Running the Service

#### 1. Clone and configure:

```bash
git clone https://github.com/your-org/caching-service.git
cd caching-service
```

Edit the `.env` file with your Redis connection details:

```yaml
redis:
  host: redis-cluster.default.svc.cluster.local
  port: 6379
  tls_enabled: true

auth:
  api_key: your-api-key
```

#### 2. Start the service

```bash
go run main.go
```

#### 3. Deploy to Kubernetes (optional)

```bash
kubectl apply -f k8s/
```

---

## 🔗 Dependencies

| Component     | Purpose                             |
| ------------- | ----------------------------------- |
| Redis         | Centralized cache store             |
| gRPC          | Communication between services      |
| OpenTelemetry | Observability (metrics/traces/logs) |
| TLS           | Secure service communication        |

---

## 📡 gRPC APIs

Service defined in: `redis.cache.v1.CacheService`

### 1. `SetCache`

#### Request

```json
{
  "namespace": "serviceA",
  "key": "user:123",
  "value": "{\"name\":\"John\"}",
  "ttl": 300,
  "tracking_id": "req-abc-123"
}
```

| Field         | Type   | Description                           |
| ------------- | ------ | ------------------------------------- |
| `namespace`   | string | Logical grouping (e.g., service name) |
| `key`         | string | Cache key                             |
| `value`       | string | Serialized JSON or plain string       |
| `ttl`         | int64  | Expiry in seconds                     |
| `tracking_id` | string | Optional for trace/debug              |

#### Response

```json
{
  "success": true,
  "message": "Cache set successfully"
}
```

---

### 2. `GetCache`

#### Request

```json
{
  "namespace": "serviceA",
  "key": "user:123",
  "tracking_id": "req-abc-123"
}
```

#### Response

```json
{
  "value": "{\"name\":\"John\"}",
  "found": true,
  "message": "Cache hit",
  "error": ""
}
```

| Field   | Type   | Description           |
| ------- | ------ | --------------------- |
| `found` | bool   | Indicates hit or miss |
| `error` | string | Error details if any  |

---

### 3. `InvalidateCache`

#### Request

```json
{
  "namespace": "serviceA",
  "key": "user:123",
  "tracking_id": "req-abc-123"
}
```

#### Response

```json
{
  "success": true,
  "message": "Key invalidated"
}
```

---

## 🧱 Redis Schema

- **Key Format**: `namespace:key`
- **Example**: `serviceA:user:123`
- TTL for auto-eviction of stale data

---

## ✨ Features

| Feature           | Description                                 |
| ----------------- | ------------------------------------------- |
| Namespacing       | Prevents cross-service key collisions       |
| TTL               | Automatic expiration of cache entries       |
| High Availability | Redis cluster + Sentinel support            |
| SDKs              | Planned for Go, Python, Node.js, Java       |
| Rate Limiting     | Per-service limits (via gateway or sidecar) |

---

## 🔐 Security

- API key-based authentication
- Role-based access: read-only / read-write
- TLS between all internal service communications

---

## 📊 Monitoring

- **Metrics**: cache hits, misses, memory, Redis latency
- **Dashboards**: Grafana with Prometheus data source
- **Alerts**: High eviction rate, Redis connection pool issues

---

## 📈 Scaling

| Component     | Scaling Method                       |
| ------------- | ------------------------------------ |
| Cache Service | Kubernetes Horizontal Pod Autoscaler |
| Redis         | Sharded Cluster + Sentinel           |

---

## 🧪 Testing

- **Unit Tests**: Validate gRPC interfaces and Redis operations
- **Load Tests**: Simulate high request volume
- **Failure Scenarios**: Redis unavailability, TTL edge cases

---

## 🧾 gRPC Proto Reference

```proto
syntax = "proto3";

package redis.cache.v1;
option go_package = "redis-service/proto";

service CacheService {
  rpc SetCache(SetCacheRequest) returns (SetCacheResponse);
  rpc GetCache(GetCacheRequest) returns (GetCacheResponse);
  rpc InvalidateCache(InvalidateCacheRequest) returns (InvalidateCacheResponse);
}

message SetCacheRequest {
  string namespace = 1;
  string key = 2;
  string value = 3;
  int64 ttl = 4;
  string tracking_id = 5;
}

message SetCacheResponse {
  bool success = 1;
  string message = 2;
}

message GetCacheRequest {
  string namespace = 1;
  string key = 2;
  string tracking_id = 3;
}

message GetCacheResponse {
  string value = 1;
  bool found = 2;
  string message = 3;
  string error = 4;
}

message InvalidateCacheRequest {
  string namespace = 1;
  string key = 2;
  string tracking_id = 3;
}

message InvalidateCacheResponse {
  bool success = 1;
  string message = 2;
}
```
