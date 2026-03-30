# Answers

## Q1
You can trace domain operations without passing `context.Context` into the domain in at least two clean ways.

### Approach 1: Trace around the domain call in the application layer
The use case owns the current span and emits events before and after the domain method:

```go
span.AddEvent("domain.order.complete.started")
err := order.Complete()
if err != nil {
    span.RecordError(err)
}
span.AddEvent("domain.order.complete.finished")
```

This is the simplest approach because the domain stays pure and the trace still shows when the domain rule executed.

### Approach 2: Use a domain event/observer port
The application layer provides a recorder or observer that translates domain events into trace span events. The domain emits plain domain events such as `OrderCompleted` or `OrderCompletionRejected`, and an outer adapter turns those events into trace annotations.

This keeps the domain infrastructure-free while still giving rich observability.

## Q2
### 1. When mocking is the right choice
Mocking is correct when the dependency is slow, nondeterministic, expensive, or irrelevant to the exact behavior under test. Example: unit testing retry logic for a payment gateway adapter by mocking the transport so you can force two timeouts and then a success.

### 2. When mocking hides bugs
Mocking hides bugs when the real integration contract matters. Example: a mocked repository says `CreateMut` succeeded, but the real SQL uses the wrong placeholder syntax for PostgreSQL or violates a `NOT NULL` constraint. The mock test passes while production fails on the first real insert.

### 3. When you need both mock test and integration test
You need both when logic exists at two levels. Example: for a payment client, use mock-based unit tests to verify retry/backoff branching and integration tests with WireMock to verify HTTP payloads, status-code handling, and JSON parsing.

## Q3
For the report "Order was charged but shows as failed", I would debug in this order:

### Logs that help
- request boundary logs for `CreateOrder`
- payment authorization logs including provider response code
- order state transition logs
- outbox publish logs if asynchronous status updates are involved
- error logs with `trace_id`, `order_id`, and failure class

### Metrics that would indicate the issue
- spike in `payment_authorizations_total{status="success"}`
- simultaneous spike in `orders_created_total{status="failure"}`
- increased `outbox_backlog_count` or message publish failures
- elevated `order_processing_duration_seconds{step="fulfillment",result="failure"}`

### Alerts
- alert when payment successes increase while completed orders drop below expected ratio
- alert on stuck pending/failed order gauge above threshold
- alert on outbox backlog age or count
- alert on payment success to order success ratio divergence over a rolling window

## Q4
Test the outbox end-to-end by removing uncontrolled background timing from the test.

Recommended flow:
1. Run a real database.
2. Execute the command that writes the aggregate and outbox entry in one transaction.
3. Assert both rows exist after commit.
4. Invoke the outbox worker synchronously from the test, or expose a `RunOnce(ctx)` method.
5. Use a stub publisher that records published messages deterministically.
6. Assert the message payload, publish count, and processed marker in the database.

This is end-to-end enough to validate transactional behavior and message publication, but deterministic because the test controls exactly when the worker runs.

## Q5
The test is wrong because it compares two independent calls to `time.Now()`. They are almost never equal, and CI timing differences make that even more obvious.

Fix it by controlling time:

1. Inject a clock interface, for example `Now() time.Time`.
2. In tests, use a fixed fake time.
3. If exact equality is not required, assert within a tolerance instead of equality.

Example:

```go
fixed := time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC)
clock := fakeClock{now: fixed}
order.CreatedAt = clock.Now()
assert.Equal(t, fixed, order.CreatedAt)
```
