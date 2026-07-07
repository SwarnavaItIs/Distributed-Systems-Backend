# DMB Benchmark Summary

| Test | Workload | Main Result |
|---|---|---|
| Search smoke test | 1 VU, 5 iterations | p95 108.32 ms, 0% failures |
| Search load test | Up to 20 VUs | 59.07 searches/sec, p95 10.69 ms |
| PostgreSQL vs Redis | 10 VUs per path | Redis p95 20.84% lower |
| Rate-limit burst | 50 concurrent requests | Exactly 20 allowed and 30 blocked |
| WebSocket sessions | 500 clients | 500 completed, 0 failed |

## Important Notes

- Search load testing used a warmed Redis cache.
- PostgreSQL and Redis comparison used unique filters for database misses and repeated filters for cache hits.
- WebSocket testing confirmed 500 stable connections but did not independently verify that every client received the listing event.
- Results were produced on a local Windows machine using Docker Desktop.