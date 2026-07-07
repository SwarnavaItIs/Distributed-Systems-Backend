# Week 4 Day 1 — Search Smoke-Test Baseline

## Environment

- Operating system: Windows
- Runtime: Docker Desktop
- API Gateway: Docker container
- Search Service: Docker container
- PostgreSQL: Docker container
- Redis: Docker container
- Virtual users: 1
- Iterations: 5
- Test duration: 5.2 seconds

## Test Route

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
| Checks passed | 100.00% |
| Checks succeeded | 20 / 20 |
| Checks failed | 0 / 20 |
| HTTP failure rate | 0.00% |
| Total HTTP requests | 6 |
| Completed iterations | 5 |
| Average HTTP latency | 28.24 ms |
| Median / p50 HTTP latency | 6.70 ms |
| p90 HTTP latency | 72.75 ms |
| p95 HTTP latency | 102.39 ms |
| Maximum HTTP latency | 132.03 ms |
| Average search latency | 32.59 ms |
| Median / p50 search latency | 6.91 ms |
| p90 search latency | 84.61 ms |
| p95 search latency | 108.32 ms |
| Maximum search latency | 132.03 ms |
| PostgreSQL responses | 1 |
| Redis cache responses | 4 |
| Data received | 5.6 kB |
| Data sent | 2.0 kB |

## Threshold Results

| Threshold | Result |
|---|---|
| Checks rate = 100% | PASS |
| HTTP failure rate = 0% | PASS |
| Search p95 < 500 ms | PASS |

## Functional Checks

All functional checks passed:

- Search status was `200`
- Response contained a `results` array
- Response contained a numeric `count`
- Response source was either `postgres` or `redis_cache`

## Cache Behaviour

Out of the five search iterations:

- 1 request was served by PostgreSQL
- 4 requests were served by Redis cache

This confirms that the first request populated the cache and subsequent repeated requests were served from Redis.

## Observations

- The smoke test completed successfully with zero failed HTTP requests.
- Search p95 latency was `108.32 ms`, comfortably below the `500 ms` threshold.
- Median search latency was only `6.91 ms`, indicating that most requests were fast.
- The maximum observed search latency was `132.03 ms`.
- Redis handled 80% of the search iterations after the initial PostgreSQL response.
- This was a low-load functional smoke test using one virtual user, so it should not yet be treated as a maximum-capacity benchmark.

## Raw Summary

```text
checks_total: 20
checks_succeeded: 100.00%
checks_failed: 0.00%

search_cache_hits: 4
search_postgres_hits: 1

http_req_duration:
  avg: 28.24 ms
  med: 6.70 ms
  p90: 72.75 ms
  p95: 102.39 ms
  max: 132.03 ms

search endpoint duration:
  avg: 32.59 ms
  med: 6.91 ms
  p90: 84.61 ms
  p95: 108.32 ms
  max: 132.03 ms

http_req_failed: 0.00%
http_reqs: 6
iterations: 5
virtual users: 1
test duration: 5.2 seconds
```

## Conclusion

The Week 4 Day 1 baseline passed all correctness and latency thresholds. The API Gateway, JWT middleware, rate limiter, Search Service, PostgreSQL, and Redis cache worked together successfully under the smoke-test workload.
