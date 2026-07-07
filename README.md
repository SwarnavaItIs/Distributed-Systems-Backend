# DMB — Distributed Marketplace Backend

DMB is a distributed C2C marketplace backend built with Go, gRPC, PostgreSQL, Redis, WebSocket, JWT authentication, and Docker Compose.

The project demonstrates service-to-service communication, caching, rate limiting, real-time event delivery, graceful shutdown, database optimization, and performance testing.

---

## Architecture

```text
                         ┌─────────────────────┐
                         │       Client        │
                         └──────────┬──────────┘
                                    │ HTTP
                                    ▼
                         ┌─────────────────────┐
                         │     API Gateway     │
                         │      Port 8080      │
                         │                     │
                         │ JWT Authentication  │
                         │ Redis Rate Limiter  │
                         └───────┬───────┬─────┘
                                 │       │
                        gRPC     │       │ HTTP proxy
                                 │       │
                                 ▼       ▼
                    ┌────────────────┐  ┌────────────────┐
                    │ Listing Service│  │ Search Service │
                    │   Port 50051   │  │   Port 8081    │
                    └───────┬────────┘  └───────┬────────┘
                            │                   │
                            │                   ├── Redis cache
                            │                   │
                            ▼                   ▼
                    ┌────────────────────────────────────┐
                    │             PostgreSQL             │
                    │          Source of truth           │
                    └────────────────────────────────────┘

                            Listing created
                                  │
                                  ▼
                         Redis Pub/Sub
                       channel: listing.created
                                  │
                                  ▼
                    ┌─────────────────────────┐
                    │  Notification Service   │
                    │       Port 8082         │
                    │                         │
                    │ WebSocket connection    │
                    │ manager + heartbeat     │
                    └────────────┬────────────┘
                                 │
                                 ▼
                       Connected WebSocket clients
```

---

## Services

### API Gateway

The Gateway is the public entry point of DMB.

Responsibilities:

- HTTP routing
- JWT HS256 validation
- Redis Lua sliding-window rate limiting
- HTTP reverse proxy to Search Service
- HTTP-to-gRPC translation for Listing Service
- Request timeouts
- Graceful shutdown

Default port:

```text
8080
```

### Listing Service

The Listing Service owns listing creation and retrieval.

Responsibilities:

- `CreateListing` gRPC method
- `GetListing` gRPC method
- Listing validation
- PostgreSQL persistence
- Publishing `listing.created` events to Redis

Default port:

```text
50051
```

### Search Service

The Search Service provides marketplace search by category and price range.

Responsibilities:

- Search filtering
- PostgreSQL query execution
- Redis cache-aside strategy
- Configurable cache TTL
- HTTP JSON responses
- Graceful shutdown

Default port:

```text
8081
```

### Notification Service

The Notification Service delivers real-time marketplace events.

Responsibilities:

- WebSocket upgrade handling
- Concurrent connection management using `sync.Map`
- Per-client read and write pumps
- Ping/Pong heartbeat
- Redis Pub/Sub subscription
- Fan-out to connected clients
- WebSocket cleanup during shutdown

Default port:

```text
8082
```

---

## Technology Stack

| Category | Technology |
|---|---|
| Language | Go |
| API Gateway | Go `net/http` |
| Internal RPC | gRPC and Protocol Buffers |
| Database | PostgreSQL |
| Cache | Redis |
| Rate limiting | Redis sorted sets and Lua |
| Real-time communication | WebSocket |
| Event transport | Redis Pub/Sub |
| Authentication | JWT HS256 |
| Containerization | Docker and Docker Compose |
| HTTP load testing | k6 |
| WebSocket testing | Artillery |

---

## Main Request Flows

### Create Listing

```text
Client
  ↓
API Gateway
  ↓
JWT middleware
  ↓
Redis rate limiter
  ↓
Gateway Listing HTTP handler
  ↓
Listing Service through gRPC
  ↓
PostgreSQL INSERT
  ↓
Redis PUBLISH listing.created
  ↓
Notification Service
  ↓
WebSocket clients
```

### Search Listings

```text
Client
  ↓
API Gateway
  ↓
JWT middleware
  ↓
Redis rate limiter
  ↓
Search Service
  ↓
Redis GET
  ├── Cache hit → return cached data
  └── Cache miss
          ↓
      PostgreSQL query
          ↓
      Redis SET with TTL
          ↓
      Return response
```

---

## API Endpoints

### Gateway

| Method | Endpoint | Description | Authentication |
|---|---|---|---|
| GET | `/health` | Gateway health check | No |
| GET | `/api/search` | Search listings | Yes |
| POST | `/api/listings` | Create listing | Yes |
| GET | `/api/listings/{id}` | Get listing by ID | Yes |

### Search Service

| Method | Endpoint | Description |
|---|---|---|
| GET | `/health` | Search Service health |
| GET | `/search` | Search listings |

Search parameters:

```text
category_id
min_price
max_price
limit
```

### Notification Service

| Method | Endpoint | Description |
|---|---|---|
| GET | `/health` | Service health and active clients |
| GET | `/ws` | WebSocket connection |
| POST | `/broadcast` | Manual development broadcast |

---

## gRPC Contract

The Listing Service exposes:

```text
listing.v1.ListingService/CreateListing
listing.v1.ListingService/GetListing
```

The contract is located at:

```text
proto/listing/v1/listing.proto
```

Generate Go stubs:

```powershell
protoc --proto_path=proto --go_out=. --go_opt=module=github.com/swarnava/dmb --go-grpc_out=. --go-grpc_opt=module=github.com/swarnava/dmb proto/listing/v1/listing.proto
```

---

## Database Design

The `listings` table contains:

```text
id
seller_id
title
description
category_id
price_cents
status
created_at
updated_at
```

Prices are stored as integer cents using `BIGINT` to avoid floating-point precision problems.

Important composite index:

```sql
CREATE INDEX idx_listings_category_price
ON listings (category_id, price_cents);
```

This index supports a common marketplace query:

```sql
SELECT *
FROM listings
WHERE category_id = $1
AND price_cents BETWEEN $2 AND $3
ORDER BY price_cents
LIMIT $4;
```

---

## Redis Usage

DMB uses Redis for three separate purposes.

### Search Cache

```text
search:<hash-of-filters>
```

Search results are stored as JSON with a short TTL.

### Sliding-Window Rate Limiter

```text
rate_limit:user:<user-id>
```

Each key is a sorted set containing accepted request timestamps.

The Lua script atomically:

1. Removes expired timestamps.
2. Counts active requests.
3. Rejects requests over the limit.
4. Inserts accepted request timestamps.
5. Refreshes key expiration.

### Pub/Sub Events

Channel:

```text
listing.created
```

Listing Service publishes events and Notification Service subscribes to them.

---

## Running the Project

### Requirements

- Docker Desktop
- Go
- PowerShell
- Optional: Node.js for `wscat`

Create local environment file:

```powershell
Copy-Item .env.example .env
```

Start the complete system:

```powershell
docker compose up --build -d
```

Check services:

```powershell
docker compose ps
```

Follow logs:

```powershell
docker compose logs -f
```

Stop the system:

```powershell
docker compose down
```

Do not use `docker compose down -v` unless you intentionally want to remove PostgreSQL data.

---

## Local URLs

```text
API Gateway:          http://localhost:8080
Search Service:       http://localhost:8081
Notification Service: http://localhost:8082
Listing gRPC:         localhost:50051
WebSocket:            ws://localhost:8082/ws
```

---

## Testing

### Generate JWT

```powershell
$TOKEN = (go run services/gateway/cmd/token/main.go).Trim()
```

### Create Listing

```powershell
Set-Content -Path listing.json -Value '{"seller_id":"11111111-1111-1111-1111-111111111111","title":"DMB Listing","description":"Marketplace listing","category_id":7,"price_cents":499900}'
```

```powershell
curl.exe -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" --data-binary "@listing.json" http://localhost:8080/api/listings
```

### Search Listings

```powershell
curl.exe -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/search?category_id=7&min_price=100000&max_price=1000000&limit=5"
```

### Connect WebSocket Client

```powershell
npx wscat -c ws://localhost:8082/ws
```

### Run End-to-End Test

```powershell
go run tests/e2e/realtime/main.go
```

---

## Performance Results

All results were recorded on a local Windows machine using Docker Desktop.

### Search Smoke Test

| Metric | Result |
|---|---:|
| Virtual users | 1 |
| Search iterations | 5 |
| Checks passed | 100% |
| HTTP failures | 0% |
| Search p95 | 108.32 ms |
| PostgreSQL responses | 1 |
| Redis responses | 4 |

### Search Load Test

| Metric | Result |
|---|---:|
| Peak virtual users | 20 |
| Search iterations | 5,323 |
| Throughput | 59.07 searches/sec |
| Successful searches | 100% |
| HTTP failures | 0% |
| Median latency | 5.44 ms |
| Search p95 | 10.69 ms |
| Redis cache hits | 5,323 |

### PostgreSQL vs Redis

| Metric | PostgreSQL | Redis |
|---|---:|---:|
| Requests | 1,440 | 1,401 |
| Success rate | 100% | 100% |
| Median latency | 7.20 ms | 6.06 ms |
| p90 latency | 11.53 ms | 9.61 ms |
| p95 latency | 13.77 ms | 10.90 ms |
| Maximum latency | 56.33 ms | 1.12 s |

Redis reduced p95 search latency by approximately:

```text
20.84%
```

The Redis run also contained a rare latency outlier, so the project reports both typical latency improvements and the observed maximum honestly.

### Rate-Limiter Concurrency Test

| Metric | Result |
|---|---:|
| Concurrent requests | 50 |
| Configured limit | 20 |
| Allowed requests | 20 |
| Blocked requests | 30 |
| Unexpected responses | 0 |
| Checks passed | 100% |
| HTTP failures | 0% |

The rate limiter produced zero over-limit escapes.

### WebSocket Connection Test

| Metric | Result |
|---|---:|
| WebSocket clients created | 500 |
| Clients completed | 500 |
| Failed clients | 0 |
| Connection lifetime | Approximately 60 seconds |

The test confirms stable handling of 500 WebSocket sessions. The supplied Artillery summary did not contain message-delivery counters, so the project does not claim verified delivery to all 500 clients.

Detailed benchmark reports are stored in:

```text
benchmarks/
```

---

## Graceful Shutdown

All Go services handle `SIGINT` and `SIGTERM`.

During shutdown:

- HTTP servers stop accepting new requests.
- Active HTTP requests receive time to finish.
- Listing Service calls `grpc.GracefulStop()`.
- Redis clients are closed.
- PostgreSQL pools are closed.
- Redis subscriptions are cancelled.
- Active WebSocket clients receive close frames.
- Docker waits using `stop_grace_period`.

---

## Consistency and Availability Trade-Offs

### Listing Service

Listing creation prioritizes strong consistency.

```text
PostgreSQL write must succeed
before the listing is returned
```

PostgreSQL is the source of truth.

### Search Service

Search allows eventual consistency.

```text
Redis may temporarily return cached data
until the cache TTL expires
```

Slightly stale search results are acceptable because search performance and availability are prioritized.

### Notification Service

Redis Pub/Sub provides low-latency delivery but does not persist events.

If Notification Service is disconnected when an event is published, that event is lost.

For durable delivery, future versions could use:

- Redis Streams
- Apache Kafka
- RabbitMQ
- Transactional outbox pattern

---

## Design Decisions

### Why gRPC for Listing Service?

- Typed contracts
- Protocol Buffer serialization
- HTTP/2 multiplexing
- Efficient internal service communication
- Generated client and server code

### Why HTTP for Search Service?

Search is easy to expose through query parameters and HTTP reverse proxying.

### Why PostgreSQL?

- Strong consistency
- Transactions
- Indexing
- Structured marketplace data
- Reliable source of truth

### Why Redis?

Redis supports several low-latency use cases:

- Cached search responses
- Atomic rate limiting
- Real-time event transport

### Why `sync.Map`?

WebSocket connections are registered, read, written, and removed concurrently by multiple goroutines.

### Why Cache-Aside?

The application controls cache population:

```text
Read Redis
    ↓ miss
Read PostgreSQL
    ↓
Write Redis
```

PostgreSQL remains authoritative.

---

## Current Limitations

- No dedicated User/Auth microservice
- JWT tokens are generated by a development helper
- Redis Pub/Sub events are not durable
- Search pagination is limited
- No transactional outbox
- No distributed tracing backend
- No verified 500-client event-delivery count
- No production Kubernetes deployment
- Benchmarks were executed on a local Docker Desktop environment

---

## Future Improvements

- Dedicated User Service
- Refresh tokens and password authentication
- Transactional outbox
- Redis Streams or Kafka
- OpenTelemetry tracing
- Prometheus and Grafana monitoring
- Circuit breaker for downstream services
- Retry policy with exponential backoff
- PostgreSQL partitioning
- Kubernetes deployment
- Service replicas and load balancing
- WebSocket delivery acknowledgements

---

## Repository Highlights

This project demonstrates:

- Microservice architecture
- API Gateway pattern
- gRPC communication
- Redis caching
- Atomic distributed rate limiting
- PostgreSQL indexing
- WebSocket connection management
- Event-driven architecture
- Docker Compose orchestration
- Graceful shutdown
- End-to-end testing
- Performance benchmarking