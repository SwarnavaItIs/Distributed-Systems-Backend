# Week 4 Day 5 — WebSocket Fan-Out Load Test

## Environment

- Runtime: Docker Desktop
- Load generator: Artillery Docker container
- Notification Service: Docker container
- Redis: Docker container
- Listing Service: Docker container
- API Gateway: Docker container
- Target WebSocket clients: 500
- Connection arrival rate: 25 clients/second
- Arrival duration: 20 seconds
- Connection lifetime: 60 seconds
- Total test duration: 1 minute 20 seconds

## Test Flow

```text
500 WebSocket clients connect
        ↓
Notification Service stores active connections
        ↓
Clients remain connected for 60 seconds
        ↓
Connections close
        ↓
Notification Service removes disconnected clients
```

## Results

| Metric | Result |
|---|---:|
| Virtual users created | 500 |
| Virtual users completed | 500 |
| Virtual users failed | 0 |
| Client creation success rate | 100.00% |
| Client completion rate | 100.00% |
| Minimum session length | 60,006.9 ms |
| Maximum session length | 60,121.7 ms |
| Mean session length | 60,012.4 ms |
| Median session length | 60,495.1 ms |
| p95 session length | 60,495.1 ms |
| p99 session length | 60,495.1 ms |
| Total test duration | 1 minute 20 seconds |

## Connection Ramp

| Reporting period | Clients created |
|---|---:|
| Initial period | 15 |
| Main ramp period | 250 |
| Final ramp period | 235 |
| Total | 500 |

## Completion Distribution

| Reporting period | Clients completed | Failed |
|---|---:|---:|
| First completion group | 15 | 0 |
| Second completion group | 250 | 0 |
| Final completion group | 235 | 0 |
| Total | 500 | 0 |

## Delivery Verification

| Check | Result |
|---|---|
| 500 clients created | PASS |
| 500 clients completed | PASS |
| Failed virtual users = 0 | PASS |
| Connections remained open for about 60 seconds | PASS |
| All clients received welcome message | Not shown in supplied Artillery summary |
| All clients received `listing.created` event | Not shown in supplied Artillery summary |
| Connected-client count returned to 0 | Not supplied |

## Important Limitation

The supplied Artillery summary reports connection lifecycle metrics, but it does not include:

```text
websocket.messages_received
websocket.messages_sent
websocket.receive_rate
```

Therefore, this run proves that the Notification Service handled 500 WebSocket client sessions without failures, but it does not by itself prove that all 500 clients received the `listing.created` broadcast.

To claim complete fan-out delivery, also record one of the following:

```text
websocket.messages_received = approximately 1000
```

where:

```text
500 welcome messages
+
500 listing.created messages
=
1000 received messages
```

or add explicit message-receipt tracking to the Artillery scenario.

## Observations

- All 500 virtual users were created successfully.
- All 500 virtual users completed successfully.
- No virtual user failed.
- Connections remained active for approximately 60 seconds as configured.
- Session duration was consistent across the full client population.
- The Notification Service successfully handled the connection lifecycle for 500 clients.
- Message-delivery metrics were not present in the supplied summary, so broadcast delivery should be verified separately.

## Raw Summary

```text
vusers.created: 500
vusers.completed: 500
vusers.failed: 0

session length:
  min: 60006.9 ms
  max: 60121.7 ms
  mean: 60012.4 ms
  median: 60495.1 ms
  p95: 60495.1 ms
  p99: 60495.1 ms

total runtime: 1 minute 20 seconds
```

## Conclusion

The Notification Service successfully handled 500 WebSocket client sessions with a 100% completion rate and zero failed virtual users. Each connection remained active for approximately 60 seconds, demonstrating stable connection management at the tested scale.

The current result validates concurrent connection handling and cleanup behavior. A separate message-delivery measurement is still required before claiming that one `listing.created` event was successfully delivered to all 500 connected clients.
