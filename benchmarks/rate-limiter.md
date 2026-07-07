# Week 4 Day 4 — Redis Rate-Limiter Concurrency Test

## Environment

- Runtime: Docker Desktop
- Load generator: k6 Docker container
- API Gateway: Docker container
- Redis: Docker container
- Concurrent virtual users: 50
- Protected requests: 50
- Total HTTP requests including setup health check: 51
- Configured rate limit: 20 requests
- Configured window: 15 seconds
- Rate-limit identity: JWT `sub` claim

## Test Design

```text
50 concurrent requests from one authenticated user
        ↓
Gateway JWT middleware
        ↓
Redis Lua sliding-window rate limiter
        ↓
20 requests allowed
30 requests blocked with HTTP 429
```

The test used 50 virtual users and 50 shared iterations, creating a short concurrent burst against the same user-specific Redis rate-limit key.

## Results

| Metric | Result |
|---|---:|
| Protected requests | 50 |
| Total HTTP requests | 51 |
| Allowed responses | 20 |
| Blocked responses | 30 |
| Unexpected responses | 0 |
| Checks passed | 100.00% |
| Checks succeeded | 150 / 150 |
| Checks failed | 0 |
| HTTP failure rate | 0.00% |
| Completed iterations | 50 |
| Interrupted iterations | 0 |
| Peak virtual users | 50 |
| Burst request rate | 183.65 req/s |
| Average burst latency | 138.33 ms |
| Median / p50 burst latency | 123.77 ms |
| p90 burst latency | 201.20 ms |
| p95 burst latency | 204.07 ms |
| Maximum burst latency | 206.65 ms |

## Expected Distribution

| Decision | Expected | Actual | Result |
|---|---:|---:|---|
| Allowed | 20 | 20 | PASS |
| Blocked | 30 | 30 | PASS |
| Unexpected | 0 | 0 | PASS |

## Functional Checks

All 150 checks passed:

- Every response status was either `200` or `429`
- Every allowed response contained a valid search payload
- Every blocked response contained the expected `rate limit exceeded` error

## Threshold Results

| Threshold | Result |
|---|---|
| Checks rate = 100% | PASS |
| Expected HTTP failure rate = 0% | PASS |
| Allowed responses = 20 | PASS |
| Blocked responses = 30 | PASS |
| Unexpected responses = 0 | PASS |

## HTTP Metrics

| Metric | Result |
|---|---:|
| Average request duration | 135.70 ms |
| Median request duration | 123.75 ms |
| p90 request duration | 200.89 ms |
| p95 request duration | 204.07 ms |
| Maximum request duration | 206.65 ms |
| HTTP request rate | 187.32 req/s |
| Average connection time | 23.05 ms |
| Average waiting time | 135.42 ms |

The scenario-only request duration was slightly higher than the overall figure because the overall metrics also included the setup health-check request.

## Redis Verification

The k6 output proves that the limiter made exactly 20 allow decisions and 30 block decisions.

Direct Redis verification should be recorded separately using:

```powershell
docker exec dmb_redis redis-cli ZCARD "rate_limit:user:11111111-1111-1111-1111-111111111111"
```

Expected value:

```text
20
```

Actual Redis sorted-set size:

```text
Not supplied with the k6 summary
```

## Window Reset Verification

Recommended verification:

```powershell
Start-Sleep -Seconds 16
docker exec dmb_redis redis-cli EXISTS "rate_limit:user:11111111-1111-1111-1111-111111111111"
```

Expected result:

```text
0
```

Then send another authenticated request. It should return HTTP `200`.

Actual window-reset result:

```text
Not supplied with the k6 summary
```

## Observations

- The limiter enforced the configured limit exactly under concurrent load.
- No over-limit request escaped.
- No unexpected status code was returned.
- All intentionally blocked requests were recognized as expected responses rather than test failures.
- The burst completed with zero interrupted iterations.
- The system processed 50 protected requests in roughly 0.3 seconds.
- The latency was much higher than the cached search benchmark because 50 requests arrived nearly simultaneously and contended for network connections, Gateway processing, and Redis rate-limit evaluation.

## Raw Summary

```text
checks_total: 150
checks_succeeded: 100.00%
checks_failed: 0.00%

allowed_responses: 20
blocked_responses: 30
unexpected_responses: 0

protected HTTP requests: 50
total HTTP requests: 51
HTTP failure rate: 0.00%

scenario request duration:
  avg: 138.33 ms
  min: 89.86 ms
  med: 123.77 ms
  p90: 201.20 ms
  p95: 204.07 ms
  max: 206.65 ms

iterations: 50
interrupted iterations: 0
peak VUs: 50
scenario throughput: 183.65 req/s
```

## Conclusion

The Redis Lua sliding-window rate limiter passed the concurrent correctness test.

With 50 simultaneous authenticated requests and a configured limit of 20, the Gateway allowed exactly 20 requests, blocked exactly 30 requests with HTTP `429`, and produced zero unexpected responses. This demonstrates that the Lua-based limiter enforced the request quota atomically without race-condition escapes under the tested burst workload.
