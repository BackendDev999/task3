# Tracing Design

## Architecture Context

The original submission focused mainly on `observability` and `testing`, which made the repository look too thin and exercise-oriented. To make the design read like a real backend service, the corrected structure now includes:

- `domain` for core business entities
- `contracts` for dependency boundaries
- `usecases` for application orchestration
- `services` for application-facing coordination
- `handlers/http` for transport entry points
- `config` for runtime configuration
- `infrastructure` for concrete adapters
- `app` for composition and dependency wiring

That structure gives tracing a clearer place in the system: transport starts traces, services/usecases orchestrate spans, and infrastructure propagates them to dependencies.

## Decision

### 1. Which option for usecases? Why?
Use **Option C: Middleware/Decorator wrapper**, while still passing `context.Context` as the standard first parameter.

This keeps usecase signatures clean:

```go
func (uc *Interactor) Execute(ctx context.Context, req *Request) (*Response, error)
```

The decorator starts and finishes spans around the use case, adds request and result attributes, and records failures. This preserves Clean Architecture because tracing remains an infrastructure concern layered around the application core rather than embedded inside business logic.

Why not Option A alone? Because reading spans directly inside every use case scatters instrumentation code across the application layer and turns tracing into repetitive boilerplate.  
Why not Option B? Because explicit `TraceContext` parameters leak observability concerns into every use case signature and encourage the same leak into deeper layers.

### 2. Which option for domain methods? Why?
Use **none of the three inside the domain itself**. Domain methods must stay pure and should not accept `context.Context`, `trace.TraceContext`, logger instances, or tracing wrappers.

Correct domain example:

```go
func (o *Order) Complete() error
```

The domain is traced from the outside by emitting span events before and after domain operations in the use case or decorator.

### 3. Which option for repository methods?
Use **Option A semantics with standard `context.Context`** on repository methods:

```go
Retrieve(ctx context.Context, orderID string) (*domain.Order, error)
Save(ctx context.Context, order *domain.Order) error
```

Repository implementations are infrastructure-facing and should propagate context into SQL drivers, HTTP clients, and downstream RPC calls. That is the correct place to bind trace context to actual I/O.

### 4. How do you trace a domain method without passing context to it?
There are two clean approaches:

1. Surround the domain call with span events in the use case:
   - `span.AddEvent("domain.order.complete.started")`
   - call `order.Complete()`
   - `span.AddEvent("domain.order.complete.finished")`

2. Use a domain observer/recorder port created in the application layer:
   - the use case creates a recorder backed by the current span
   - the recorder emits events before and after domain execution
   - the domain still receives no infrastructure dependency

The simplest default is approach 1.

## Recommended Propagation Model

### Entry points
- API gateway injects W3C Trace Context headers such as `traceparent`.
- The HTTP transport extracts headers and starts a server span.
- The handler passes `ctx` into the service/usecase boundary.

### Application layer
- Usecase decorators create child spans such as `CreateOrderUsecase`.
- Usecase code can add semantic attributes like `order.id`, `customer.tier`, and `payment.method`.
- The same `ctx` flows to repositories and outbound clients.

### Infrastructure layer
- Repository methods use `ctx` for SQL queries.
- Payment and inventory clients inject the current trace context into outbound HTTP or gRPC calls.
- Failures are recorded on the current span and mapped back into typed application errors.

## Example Implementation

The implementation uses a tracing decorator around the interactor:

- [observability/tracing.go](/Users/yusofzaky/Documents/dev/answer/task3/observability/tracing.go)

Key points:
- The decorator owns span lifecycle.
- The usecase signature remains clean.
- Repositories consume `context.Context`.
- Domain methods remain pure.

## Example Flow

1. HTTP middleware extracts trace headers and starts request span.
2. Service calls traced usecase with the request `ctx`.
3. Decorator creates `CreateOrder` child span.
4. Usecase loads order/inventory/payment data through repositories and clients using the same `ctx`.
5. Around `order.Complete()`, the use case records span events to trace domain behavior without polluting the entity.
6. Outbound calls automatically join the same distributed trace.
