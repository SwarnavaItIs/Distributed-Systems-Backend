# Week 4 Day 2 — Search Load-Test Results

## Environment

- Operating system: Windows
- Runtime: Docker Desktop
- Load generator: k6 Docker container
- API Gateway: Docker container
- Search Service: Docker container
- PostgreSQL: Docker container
- Redis: Docker container
- Peak virtual users: 20
- Test duration: 1 minute 30.1 seconds
- Completed iterations: 5,323
- Interrupted iterations: 0
- Search cache: pre-warmed before the load phase

## Load Profile

| Stage | Duration | Target VUs |
|---|---:|---:|
| Ramp up | 10 seconds | 5 |
| Ramp up | 20 seconds | 10 |
| Ramp up | 30 seconds | 20 |
| Sustained load | 20 seconds | 20 |
| Ramp down | 10 seconds | 0 |

## Endpoint

```text
GET /api/search
```

Query parameters:

```text
category_id=7
min_price=100000
max_price=1000000
limit=5
```

## Results

| Metric | Result |
|---|---:|
| Total HTTP requests | 5,325 |
| Requests per second | 59.09 req/s |
| Completed search iterations | 5,323 |
| Iterations per second | 59.07 iter/s |
| Checks passed | 100.00% |
| Checks succeeded | 21,292 / 21,292 |
| Checks failed | 0 / 21,292 |
| Successful searches | 100.00% |
| HTTP failure rate | 0.00% |
| Average search latency | 6.16 ms |
| Median / p50 search latency | 5.44 ms |
| p90 search latency | 8.95 ms |
| p95 search latency | 10.69 ms |
| p99 search latency | Not reported in the supplied summary |
| Maximum search latency | 147.87 ms |
| Redis cache hits | 5,323 |
| Measured PostgreSQL responses | 0 |
| Rate-limited responses | 0 observed |
| Peak virtual users | 20 |
| Data received | 5.8 MB |
| Data sent | 2.0 MB |

## Functional Checks

All functional checks passed:

- Search response status was `200`
- Response contained a `results` array
- Response contained a numeric `count`
- Response source was either `postgres` or `redis_cache`

## Cache Behaviour

All 5,323 measured search iterations were served from Redis cache.

```text
Redis cache hits: 5,323
Measured PostgreSQL responses: 0
```

The search cache was intentionally warmed during the setup phase, so this benchmark measures the cached search path rather than uncached PostgreSQL query performance.

## Latency Summary

| Percentile | Search latency |
|---|---:|
| Average | 6.16 ms |
| p50 / Median | 5.44 ms |
| p90 | 8.95 ms |
| p95 | 10.69 ms |
| Maximum | 147.87 ms |

Most requests completed in under 11 ms at the p95 level. A small number of slower requests increased the maximum latency to 147.87 ms, but this did not affect the overall success rate.

## Throughput

The system completed:

```text
5,323 search iterations in 90.1 seconds
```

Observed throughput:

```text
59.07 completed search iterations per second
59.09 total HTTP requests per second
```

The two additional HTTP requests came from the setup phase, which performed the Gateway health check and search-cache warmup.

## Threshold Results

| Threshold | Result |
|---|---|
| Checks rate > 99% | PASS |
| Successful searches > 99% | PASS |
| HTTP failure rate < 1% | PASS |
| Search p95 < 500 ms | PASS |
| Search p99 < 1000 ms | Not independently verifiable from the supplied summary |

## Observations

- All 21,292 functional checks passed.
- Every measured search iteration returned HTTP `200`.
- No HTTP request failures were recorded.
- No rate-limit response was observed.
- Redis served all 5,323 measured search iterations.
- Search p95 latency was 10.69 ms.
- The system sustained approximately 59 requests per second while ramping to 20 concurrent virtual users.
- No iterations were interrupted during ramp-down.
- This test measures the warmed Redis-cache path, not the uncached PostgreSQL search path.

## Raw Summary

```text
checks_total: 21,292
checks_succeeded: 100.00%
checks_failed: 0.00%

search_cache_hits: 5,323
successful_searches: 100.00%

search_latency:
  avg: 6.16 ms
  min: 2.20 ms
  med: 5.44 ms
  max: 147.87 ms
  p90: 8.95 ms
  p95: 10.69 ms

http_req_failed: 0.00%
http_reqs: 5,325
http request rate: 59.09 req/s

iterations: 5,323
iteration rate: 59.07 iter/s
peak VUs: 20
interrupted iterations: 0
test duration: 1m30.1s

data received: 5.8 MB
data sent: 2.0 MB
```

## Conclusion

The DMB cached search path remained stable under a gradual ramp to 20 concurrent virtual users. It sustained approximately 59 requests per second with a 0% HTTP failure rate, 100% successful searches, and a search p95 latency of 10.69 ms.

Because the cache was pre-warmed, these results specifically demonstrate the performance of the API Gateway, JWT validation, Redis-backed rate limiter, Search Service, and Redis cache under moderate concurrent load.
