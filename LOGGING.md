# Logging Strategy

## Decision

### 1. Which option maintains Clean Architecture?
Use **Option C: Decorator pattern**, combined with request/response logging at service or transport boundaries.

Why:
- It keeps logging out of core usecase logic.
- It avoids duplicating infrastructure concerns in every interactor.
- It allows consistent structured logging policies across all use cases.

Option A spreads logging throughout the application layer and mixes operational concerns with business orchestration.  
Option B is correct for boundary logging, but by itself it does not provide consistent usecase-level logs unless each service method is manually instrumented.

The cleanest design is:
- boundary/service layer logs inbound/outbound requests
- logging decorators wrap usecases for operational visibility
- domain layer does not log directly

### 2. Business logs vs operational logs
**Business logs** describe meaningful domain outcomes:
- order completed
- payment authorization failed
- stock reservation rejected

These logs should be sparse, intentional, and tied to business events.

**Operational logs** describe execution behavior:
- request received
- dependency timeout
- SQL retry exhausted
- downstream 500 returned from payment provider

These logs support debugging, incident response, and SRE workflows.

### 3. How do you avoid logging sensitive data?
- Redact or hash PII before logging.
- Never log secrets, tokens, passwords, full PAN, CVV, or auth headers.
- Whitelist safe fields instead of blacklisting dangerous ones.
- Keep request/response log serializers separate from domain models.
- Apply data classification rules consistently in one logging utility.

## Logging Rules

### Where logging should happen
- HTTP/gRPC boundary: request accepted, response completed, latency, status code.
- Usecase decorator: business operation started/completed/failed.
- Infrastructure adapters: dependency failures, retries, timeouts, non-success remote responses.

### Where logging should not happen
- Domain entities and value objects.
- Low-value happy-path internal helper functions.
- Every repository query by default, unless debugging or audit requirements justify it.

## Structured Logging Requirements

Every log entry should contain:
- `timestamp`
- `level`
- `message`
- `trace_id`
- `span_id`
- `service`
- `operation`
- stable business identifiers such as `order_id` when available

## Request/Response Logging at Boundaries

Inbound:
- method
- route
- request_id
- caller identity if available
- redacted request payload

Outbound:
- status code
- duration
- redacted response payload
- error classification

## Error Logging with Stack Traces

Stack traces should be logged when:
- a request fails unexpectedly
- infrastructure returns an unhandled error
- a panic is recovered

Do not attach stack traces to expected business rejections such as validation failures unless the team explicitly wants them for diagnostics.

## PII Redaction Strategy

Recommended pattern:
- define transport DTOs for logging
- map sensitive fields to `[REDACTED]`
- partially mask emails and phone numbers
- truncate large payloads

Example:

```json
{
  "customer_id": "cust_123",
  "customer_email": "j***@example.com",
  "card_token": "[REDACTED]",
  "amount_cents": 159900
}
```

## Example Implementation

- [observability/logging.go](/Users/yusofzaky/Documents/dev/answer/task3/observability/logging.go)

This implementation provides:
- a logger wrapper that injects `trace_id` and `span_id` from `context.Context`
- a logging decorator around usecases
- request redaction helpers
- error logging with stack traces

## Sample Log Events

### Request accepted
```json
{
  "level": "INFO",
  "message": "http.request.started",
  "trace_id": "7d1c8d8f6a4948c3b8a21374df267dcb",
  "route": "POST /orders",
  "request": {
    "customer_id": "cust_123",
    "customer_email": "j***@example.com",
    "card_token": "[REDACTED]"
  }
}
```

### Usecase failed
```json
{
  "level": "ERROR",
  "message": "usecase.failed",
  "trace_id": "7d1c8d8f6a4948c3b8a21374df267dcb",
  "usecase": "CreateOrder",
  "error": "payment authorize: timeout",
  "stacktrace": "goroutine 1 ..."
}
```
