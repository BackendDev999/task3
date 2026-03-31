# Task 3 Service

This directory contains a runnable Go service that demonstrates:

- layered backend structure
- observability design artifacts
- integration testing examples
- a simple HTTP entry point for order creation

## Structure

```text
task3/
├── app/
├── config/
├── contracts/
├── domain/
├── handlers/http/
├── infrastructure/
├── observability/
├── services/
├── testing/
├── usecases/
└── main.go
```

## Run

From the workspace root:

```bash
cd /Users/yusofzaky/Documents/dev/answer
GOCACHE=/tmp/go-build go run ./task3
```

Default address:

```text
:8080
```

Optional environment variables:

```bash
export HTTP_ADDRESS=:8080
export PAYMENT_BASE_URL=http://payment.local
export INVENTORY_BASE_URL=http://inventory.local
```

## Endpoints

### Health check

```bash
curl -i http://localhost:8080/health
```

Example response:

```json
{"status":"ok"}
```

### Create order

```bash
curl -i -X POST http://localhost:8080/orders \
  -H 'Content-Type: application/json' \
  -d '{
    "order_id": "ord-001",
    "customer_id": "cust-001",
    "amount_cents": 150000,
    "customer_tier": "premium"
  }'
```

Example response:

```json
{"order_id":"ord-001","status":"AUTHORIZED"}
```

## Build

```bash
cd /Users/yusofzaky/Documents/dev/answer
GOCACHE=/tmp/go-build go build -o /tmp/task3-app ./task3
```
