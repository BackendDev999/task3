# Integration Testing Strategy

## The Core Problem
The current tests over-use mocks, so they verify that code calls mocked methods, not that the system actually works. That misses bugs in:
- SQL generation
- schema constraints
- serialization and deserialization
- transaction boundaries
- HTTP contract mapping
- timeout handling

## Testing Boundaries

| Test Level | What to Test | What to Mock |
|------------|--------------|--------------|
| Unit | Pure domain rules, value objects, small orchestration branches, mapping functions, retry policies | Repositories, external APIs, clocks, UUID generators |
| Integration | Repository SQL, migrations, transaction behavior, serialization, HTTP clients against stub servers, outbox persistence | Third-party systems outside the boundary under test |
| E2E | Full request flow across real service boundaries, auth, persistence, async delivery, production-like config | Only truly external dependencies that cannot be run in the test environment |

## Recommended Strategy

### Unit tests
Use unit tests for:
- domain invariants
- pure calculation logic
- branch coverage for application orchestration
- error mapping

These tests should be fast and deterministic.

### Integration tests
Use integration tests for:
- repositories with a real database
- migrations and constraints
- transaction rollback behavior
- outbound clients against a stubbed external server such as WireMock
- outbox processing with real persistence

### E2E tests
Use E2E tests for:
- create order from public API through persistence and side effects
- trace propagation across service boundaries
- realistic failure scenarios that cross multiple processes

Keep E2E small in number and high in value.

## Repository Integration Test with Testcontainers

Goals:
- prove generated SQL runs on the real database engine
- prove schema constraints are enforced
- prove `UpdateMut` touches only dirty fields
- verify committed data exactly matches expectations

Implementation file:
- [testing/integration/order_repo_test.go](/Users/yusofzaky/Documents/dev/answer/task3/testing/integration/order_repo_test.go)

Container setup:
- [testing/setup/testcontainers.go](/Users/yusofzaky/Documents/dev/answer/task3/testing/setup/testcontainers.go)

What this catches that mocks miss:
- wrong SQL placeholders
- missing columns
- invalid migrations
- transaction isolation issues
- real scan/type conversion failures

## External Service Test with WireMock

Goals:
- verify HTTP payload shape
- verify response parsing
- verify timeout handling
- verify provider-specific error mapping

Implementation file:
- [testing/integration/payment_client_test.go](/Users/yusofzaky/Documents/dev/answer/task3/testing/integration/payment_client_test.go)

What this catches that mocks miss:
- wrong URL path
- wrong headers
- incorrect JSON field names
- timeout configuration bugs
- bad response body handling

## Anti-Flake Principles

- Control time with a fake clock
- Control IDs with deterministic generators
- Avoid `time.Sleep` in assertions
- Poll with bounded eventual assertions only when async work is unavoidable
- Run background workers in-process for tests when possible
- Assert persisted state, not implementation details

## Outbox Testing Strategy

To test the outbox pattern without flaky timing:

1. Write aggregate and outbox row in one transaction.
2. Commit.
3. Trigger the outbox worker explicitly from the test instead of waiting on background timing.
4. Use a stub message publisher.
5. Assert:
   - the message was published once
   - the outbox row is marked processed
   - retries are recorded correctly on failure

This removes race-prone sleep-based timing from the test while still exercising real persistence behavior.

## Coverage Matrix

### Unit
- `Order.Complete` rejects invalid states
- failure classification maps infrastructure errors to application errors
- redaction removes secrets

### Integration
- repository persists valid order rows
- invalid rows fail on real constraints
- payment client handles 200, 402, and timeout paths
- transaction rolls back when audit insert fails

### E2E
- `POST /orders` returns success and persists order
- payment success plus inventory failure produces the correct user-visible result
- trace ID is propagated across API gateway, order service, payment service, and inventory service
