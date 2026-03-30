# Metrics Design

## 1. Where do you instrument?
Instrument at multiple layers, but with different intent:

- **Service / transport layer**
  - request rate
  - response codes
  - request duration
  - payload size

- **Usecase layer**
  - business counters such as `orders_created_total`
  - workflow duration such as validation, payment, fulfillment
  - business backlog gauges such as pending orders

- **Repository / client layer**
  - dependency latency
  - dependency error rate
  - connection pool pressure

The primary business metrics belong in the **usecase layer** because that is where the system understands business outcomes. Infrastructure metrics alone cannot tell you whether order creation failed because of validation, payment, inventory, or orchestration logic.

## Recommended Metrics

### Counters
```text
orders_created_total{status="success|failure", customer_tier="free|premium"}
payment_authorizations_total{status="success|failure", provider="stripe"}
inventory_reservations_total{status="success|failure"}
```

### Histograms
```text
order_processing_duration_seconds{step="validation|payment|fulfillment", result="success|failure"}
dependency_request_duration_seconds{dependency="payment|inventory", operation="authorize|reserve", result="success|failure"}
```

### Gauges
```text
orders_pending_count
outbox_backlog_count
worker_inflight_jobs
```

## 2. Why this causes metric explosion and how to fix it

Problematic example:

```go
orderDuration.WithLabelValues(customerID, productID, orderID).Observe(duration)
```

This causes **high-cardinality explosion**:
- `customerID`, `productID`, and especially `orderID` create near-unbounded label sets
- Prometheus storage grows rapidly
- queries become slow
- memory usage increases
- alert quality drops

### Fix
Only use low-cardinality labels with bounded values:

```go
orderDuration.WithLabelValues(step, result).Observe(duration.Seconds())
```

Good labels:
- `status=success|failure`
- `customer_tier=free|premium|enterprise`
- `dependency=payment|inventory`
- `error_class=timeout|validation|conflict`

Bad labels:
- `order_id`
- `customer_id`
- `email`
- `product_id` if product catalog is large and dynamic

Use logs and traces for per-order diagnostics, not metrics.

## 3. How do you correlate metrics with traces?

Do not put trace IDs into metric labels because that creates the same cardinality problem. Instead:

- use **exemplars** on histogram observations where supported
- include stable operation names in both traces and metrics
- jump from an alerting metric to traces filtered by service, route, operation, and time window

Example:
- metric alert says `payment_authorizations_total{status="failure"}` spiked
- operator opens traces for `operation=CreateOrder` and `dependency=payment` during the same period
- logs with the same `trace_id` provide exact request-level details

## Instrumentation Placement

### Usecase
- count business success/failure
- measure end-to-end workflow duration
- classify failure reason into small bounded categories

### Repository and clients
- observe dependency latency
- count retries and timeouts
- expose pool/connection health

## Example Implementation

- [observability/metrics.go](/Users/yusofzaky/Documents/dev/answer/task3/observability/metrics.go)

This implementation shows:
- business counters with bounded labels
- histograms with step/result dimensions only
- a pending-order gauge
