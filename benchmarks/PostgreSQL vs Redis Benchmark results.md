# Week 4 Day 3 — PostgreSQL vs Redis Cache Benchmark

## Environment

- Operating system: Windows
- Runtime: Docker Desktop
- Load generator: k6 Docker container
- API Gateway: Docker container
- Search Service: Docker container
- PostgreSQL: Docker container
- Redis: Docker container
- Virtual users per scenario: 10
- PostgreSQL scenario duration: 30 seconds
- Redis scenario duration: 30 seconds
- Total elapsed time: 1 minute 5.3 seconds
- Completed iterations: 2,841
- Interrupted iterations: 0

## Test Design

### PostgreSQL Scenario

Each request used a unique `max_price` value, producing a different search-cache key.

```text
Unique search filters
        ↓
Redis cache miss
        ↓
PostgreSQL query
        ↓
Result stored in Redis
```

### Redis Scenario

Each request used the same pre-warmed search filters.

```text
Repeated search filters
        ↓
Redis cache hit
        ↓
Cached result returned
```

## Overall Results

| Metric | Result |
|---|---:|
| Total HTTP requests | 2,843 |
| Total completed iterations | 2,841 |
| Overall request rate | 43.56 req/s |
| Overall checks passed | 100.00% |
| Checks succeeded | 11,364 / 11,364 |
| Checks failed | 0 |
| HTTP failure rate | 0.00% |
| Peak virtual users | 20 |
| Data received | 3.1 MB |
| Data sent | 1.1 MB |

## Scenario Comparison

| Metric | PostgreSQL | Redis |
|---|---:|---:|
| Requests | 1,440 | 1,401 |
| Request rate | 22.06 req/s | 21.47 req/s |
| Success rate | 100.00% | 100.00% |
| Average latency | 8.07 ms | 14.02 ms |
| Median / p50 latency | 7.20 ms | 6.06 ms |
| p90 latency | 11.53 ms | 9.61 ms |
| p95 latency | 13.77 ms | 10.90 ms |
| p99 latency | Not reported | Not reported |
| Maximum latency | 56.33 ms | 1.12 s |
| HTTP failures | 0 | 0 |
| Completed iterations | 1,440 | 1,401 |

## Source Verification

| Metric | Result |
|---|---:|
| PostgreSQL responses | 1,440 |
| Redis responses | 1,401 |
| Unexpected response sources | 0 observed |
| Overall HTTP failure rate | 0.00% |
| Interrupted iterations | 0 |

All source-validation checks passed:

- PostgreSQL scenario returned `source: postgres`
- Redis scenario returned `source: redis_cache`
- Both scenarios returned HTTP `200`
- Both scenarios returned a valid `results` array and numeric `count`

## Latency Improvement

### p95 improvement

```text
PostgreSQL p95 - Redis p95
= 13.77 ms - 10.90 ms
= 2.87 ms
```

### p95 percentage reduction

```text
((13.77 - 10.90) / 13.77) × 100
= 20.84%
```

Redis reduced p95 latency by **2.87 ms**, or approximately **20.84%**, compared with the PostgreSQL path.

### p50 improvement

```text
PostgreSQL p50 - Redis p50
= 7.20 ms - 6.06 ms
= 1.14 ms
```

Redis reduced median latency by approximately **15.83%**.

## Threshold Results

| Threshold | Result |
|---|---|
| Checks rate > 99% | PASS |
| HTTP failure rate < 1% | PASS |
| PostgreSQL success rate > 99% | PASS |
| Redis success rate > 99% | PASS |
| PostgreSQL p95 < 500 ms | PASS |
| PostgreSQL p99 < 1000 ms | Not independently verifiable from supplied output |
| Redis p95 < 200 ms | PASS |
| Redis p99 < 500 ms | Not independently verifiable from supplied output |

## Observations

- Both scenarios completed with a 100% success rate and zero HTTP failures.
- Redis improved median latency from `7.20 ms` to `6.06 ms`.
- Redis improved p90 latency from `11.53 ms` to `9.61 ms`.
- Redis improved p95 latency from `13.77 ms` to `10.90 ms`.
- Redis p95 was approximately `20.84%` lower than PostgreSQL p95.
- PostgreSQL showed a lower average latency (`8.07 ms`) than Redis (`14.02 ms`) because the Redis scenario contained a large `1.12 s` outlier.
- Redis therefore performed better for typical requests at p50, p90, and p95, but its average and maximum were negatively affected by one or more rare slow requests.
- The benchmark validates that both response paths were measured correctly and that no unexpected source values appeared.
- No iterations were interrupted during either scenario.

## Interpretation

The Redis path was faster for most requests, as shown by its lower median, p90, and p95 values. However, its maximum latency was much higher because of a rare `1.12 s` spike.

This means the most accurate conclusion is:

```text
Redis improved typical and tail latency up to p95,
but the run also contained a significant Redis-path outlier.
```

The outlier should be investigated in a later repeat run by checking Docker CPU, memory, connection reuse, garbage collection, and host scheduling.

## Raw Summary

```text
Overall:
  checks: 11,364 / 11,364 passed
  HTTP failures: 0.00%
  HTTP requests: 2,843
  request rate: 43.56 req/s
  iterations: 2,841
  elapsed time: 1m05.3s

PostgreSQL:
  responses: 1,440
  success rate: 100.00%
  avg: 8.07 ms
  min: 2.94 ms
  med: 7.20 ms
  p90: 11.53 ms
  p95: 13.77 ms
  max: 56.33 ms
  throughput: 22.06 req/s

Redis:
  responses: 1,401
  success rate: 100.00%
  avg: 14.02 ms
  min: 2.34 ms
  med: 6.06 ms
  p90: 9.61 ms
  p95: 10.90 ms
  max: 1.12 s
  throughput: 21.47 req/s
```

## Conclusion

The controlled benchmark successfully compared uncached PostgreSQL searches with Redis-cached searches under the same virtual-user count.

Redis reduced p95 latency from `13.77 ms` to `10.90 ms`, an improvement of approximately `20.84%`. It also improved median and p90 latency. However, a `1.12 s` Redis-path outlier increased its average and maximum latency, so the result should be presented as an improvement in typical and p95 latency rather than a universal improvement across every metric.
