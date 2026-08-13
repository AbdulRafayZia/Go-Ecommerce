# GoCommerce — Distributed E-Commerce Platform in Go

**A portfolio project demonstrating 3–4 years of production-grade Go backend engineering.**

This document is the master spec. Feed it to Claude Code section by section (start with Phase 0/1) and it can scaffold, implement, and test each piece incrementally. Don't try to build everything in one shot — follow the phased roadmap at the end.

---

## 1. What this project proves to a hiring team

| Skill hiring teams screen for | Where it shows up in this project |
|---|---|
| Microservices decomposition | 6 independently deployable Go services, each owning its data |
| REST APIs + OpenAPI | All internal service-to-service calls, documented with OpenAPI/Swagger |
| GraphQL (optional) | Public-facing API at the gateway, aggregating data from multiple services |
| Event-driven architecture | Kafka for order/payment/inventory workflows, outbox pattern |
| Clean/Hexagonal architecture | Every service separates domain, application, and infrastructure layers |
| Resilience patterns | Circuit breakers, retries, idempotency, graceful shutdown |
| Observability | Structured logs, Prometheus metrics, OpenTelemetry tracing |
| CI/CD | GitHub Actions: lint → test → build → containerize → deploy |
| Containers & orchestration | Docker Compose for local dev, Kubernetes manifests for "production" |
| Go concurrency | Worker pools, context propagation, goroutine-safe caching |
| Testing discipline | Table-driven unit tests, integration tests via testcontainers-go |

---

## 2. System overview

**Domain**: A simplified e-commerce marketplace — browse products, add to cart, place an order, pay, get notified, track inventory.

**Services** (each is its own Go module, own repo folder, own database):

1. **API Gateway** — single entry point for the React frontend; handles auth (JWT), request routing, rate limiting, routes REST requests to internal services.
2. **Product Service** — catalog, search, categories. Postgres.
3. **Cart Service** — session-based cart, TTL-based. Redis.
4. **Order Service** — order lifecycle state machine, implements the **outbox pattern** to publish events reliably. Postgres.
5. **Payment Service** — mock/Stripe sandbox integration, idempotent payment processing. Postgres.
6. **Inventory Service** — stock levels, reservation on order, consumes Kafka events. Postgres.
7. **Notification Service** — consumes Kafka events, sends mock emails/webhooks. No DB (or lightweight log store).

**Cross-cutting infrastructure**:
- **Kafka** (or NATS JetStream if you want something lighter) — event bus for `order.created`, `payment.completed`, `inventory.reserved`, `order.shipped`.
- **Redis** — cart storage + rate-limiting counters + hot-path product cache.
- **PostgreSQL** — one logical database per service (can be one Postgres instance with separate schemas for local dev, separate instances for "prod").
- **Prometheus + Grafana** — metrics.
- **Jaeger + OpenTelemetry** — distributed tracing across the whole request path.
- **Docker Compose** — local orchestration of everything above.
- **Kubernetes (kind/k3d locally)** — deployment manifests + optional Helm chart.
- **GitHub Actions** — CI/CD pipeline.

---

## 3. Architecture principles

### 3.1 Hexagonal (ports & adapters) architecture — apply to every service

```
service/
├── cmd/
│   └── server/main.go          # wires everything together (composition root)
├── internal/
│   ├── domain/                 # entities, value objects, domain errors — no framework imports
│   │   ├── order.go
│   │   └── order_test.go
│   ├── application/            # use cases / services — orchestrates domain + ports
│   │   ├── create_order.go
│   │   └── create_order_test.go
│   ├── ports/                  # interfaces the application layer depends on
│   │   ├── repository.go       # e.g. OrderRepository interface
│   │   └── publisher.go        # e.g. EventPublisher interface
│   └── adapters/               # concrete implementations of ports
│       ├── postgres/           # implements repository.go
│       ├── kafka/               # implements publisher.go
│       └── http/                # REST handler layer (implements the OpenAPI spec)
├── api/                         # openapi.yaml + generated types/clients for this service
├── migrations/                  # SQL migrations (golang-migrate)
├── Dockerfile
└── go.mod
```

**Why this matters for interviews**: it lets you say "my domain layer has zero dependencies on Postgres or HTTP — I could swap either without touching business logic," which is exactly the kind of sentence senior engineers say.

### 3.2 Communication rules

- **Internal (service-to-service)**: REST/JSON over HTTP, documented with OpenAPI (Swagger) specs per service. Each service exposes a small, well-typed HTTP API; internal clients are thin Go structs generated or hand-written from the OpenAPI spec, not raw `http.Get` calls scattered everywhere.
- **External (client-to-gateway)**: REST by default. GraphQL is an optional stretch goal at the gateway layer only — it's most useful there because it lets the frontend fetch product + cart + order data in one request instead of three, which is the actual selling point of GraphQL (aggregation), not something you need between two backend services.
- **Async (workflow events)**: Kafka. Services publish domain events; other services subscribe. No service calls another synchronously just to trigger a side effect — that's what events are for.

**Note on gRPC**: The original plan used gRPC internally. REST is a fine substitute for a portfolio project and is what most companies actually use for internal service-to-service calls unless they're operating at very large scale. If you want to pick gRPC back up later, the cleanest way is to add it to *one* service (Order Service is a good candidate) once the rest of the system is working and you're comfortable with the concepts — that's a much easier way to learn it than trying to learn gRPC and the whole system at once.

### 3.3 The outbox pattern (implement this in Order Service — it's a big signal)

Problem: if Order Service writes to Postgres AND publishes to Kafka as two separate operations, a crash between them causes an inconsistent state (order saved, event lost, or vice versa).

Solution:
1. In the same DB transaction as inserting the order, insert a row into an `outbox_events` table.
2. A separate background poller (or Debezium/CDC if you want to go further) reads unpublished rows from `outbox_events` and publishes them to Kafka, then marks them published.
3. This guarantees at-least-once delivery with no dual-write problem.

```sql
CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type TEXT NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ
);
```

### 3.4 Resilience patterns to implement

| Pattern | Where | Library / approach |
|---|---|---|
| Circuit breaker | Gateway → services, Order → Payment | `sony/gobreaker` |
| Retry with backoff | All internal HTTP clients | HTTP middleware/RoundTripper wrapping `cenkalti/backoff` |
| Idempotency keys | Payment Service | Client-supplied idempotency key stored + checked before processing |
| Graceful shutdown | Every service | `context.Context` + signal handling, drain in-flight requests before exit |
| Timeouts everywhere | Every internal HTTP call | `context.WithTimeout` passed down the call chain |
| Rate limiting | API Gateway | Token bucket, per-IP or per-user, backed by Redis |

---

## 4. Data model (per service)

Use `mermaid erDiagram` syntax if you want to render this — shown here as plain schema.

**Product Service (Postgres)**
```
products(id UUID PK, name, description, price_cents, currency, category_id FK, stock_hint INT, created_at, updated_at)
categories(id UUID PK, name, parent_id FK NULLABLE)
```

**Order Service (Postgres)**
```
orders(id UUID PK, user_id UUID, status TEXT, total_cents INT, created_at, updated_at)
order_items(id UUID PK, order_id FK, product_id UUID, quantity INT, unit_price_cents INT)
outbox_events(id UUID PK, aggregate_type, aggregate_id, event_type, payload JSONB, created_at, published_at NULLABLE)
```

Order status state machine: `pending → awaiting_payment → paid → fulfilling → shipped → delivered`, with `cancelled` and `failed` branches.

**Payment Service (Postgres)**
```
payments(id UUID PK, order_id UUID, amount_cents INT, status TEXT, idempotency_key TEXT UNIQUE, provider_ref TEXT, created_at)
```

**Inventory Service (Postgres)**
```
stock(product_id UUID PK, available INT, reserved INT, updated_at)
reservations(id UUID PK, order_id UUID, product_id UUID, quantity INT, status TEXT, created_at)
```

**Cart Service (Redis)**
```
Key: cart:{user_id}   Value: JSON {items: [{product_id, quantity}], updated_at}   TTL: 7 days
```

---

## 5. REST API contract example (Order Service)

Document every service's API with an OpenAPI spec (`openapi.yaml` in each service folder). Generate Go request/response structs from it with `oapi-codegen` so your handlers and clients stay in sync with the spec — this is the REST equivalent of what `buf generate` does for proto, and it's a good thing to mention on a resume ("API-first design with generated types").

```yaml
openapi: 3.0.3
info:
  title: Order Service API
  version: 1.0.0
paths:
  /orders:
    post:
      summary: Create a new order
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateOrderRequest'
      responses:
        '201':
          description: Order created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Order'
  /orders/{orderId}:
    get:
      summary: Get an order by ID
      parameters:
        - name: orderId
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: The order
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Order'
components:
  schemas:
    CreateOrderRequest:
      type: object
      required: [user_id, items, idempotency_key]
      properties:
        user_id: { type: string, format: uuid }
        idempotency_key: { type: string }
        items:
          type: array
          items:
            $ref: '#/components/schemas/OrderItem'
    OrderItem:
      type: object
      required: [product_id, quantity]
      properties:
        product_id: { type: string, format: uuid }
        quantity: { type: integer, minimum: 1 }
    Order:
      type: object
      properties:
        id: { type: string, format: uuid }
        user_id: { type: string, format: uuid }
        status: { type: string, enum: [pending, awaiting_payment, paid, fulfilling, shipped, delivered, cancelled, failed] }
        total_cents: { type: integer }
        items:
          type: array
          items:
            $ref: '#/components/schemas/OrderItem'
```

**Idempotency**: the `idempotency_key` field on `CreateOrderRequest` matters more without gRPC's built-in retry semantics — clients (including the gateway retrying a timed-out call) must send the same key on retry, and Order Service checks it before creating a duplicate order.

**Live order status (optional stretch goal)**: instead of a gRPC streaming RPC, use **Server-Sent Events** (`GET /orders/{orderId}/events`, `Content-Type: text/event-stream`) or a WebSocket endpoint. SSE is simpler to reason about than gRPC streaming and works natively with `fetch` in the browser, so it's a good fit if you're adding the React dashboard later.

---

## 6. Repository structure (monorepo, recommended for a portfolio project)

A monorepo is easier to demo and easier for Claude Code to reason about in one session. Polyrepo is more "real-world" but adds friction for a solo portfolio project — mention in your README that you're aware of the tradeoff.

```
gocommerce/
├── services/
│   ├── gateway/
│   ├── product/
│   ├── cart/
│   ├── order/
│   ├── payment/
│   ├── inventory/
│   └── notification/
├── api/                        # shared OpenAPI specs, oapi-codegen config
├── pkg/                        # shared Go libraries (logging, tracing middleware, errors)
│   ├── logger/
│   ├── tracing/
│   ├── middleware/
│   └── httpclient/          # shared internal HTTP client with retry/circuit-breaker/tracing built in
├── deploy/
│   ├── docker-compose.yml
│   ├── k8s/
│   │   ├── base/
│   │   └── overlays/
│   └── helm/                   # optional
├── frontend/                   # optional React dashboard
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── cd.yml
├── docs/
│   ├── architecture.md
│   ├── adr/                    # architecture decision records — do this, it's a great signal
│   └── runbook.md
├── Makefile
└── README.md
```

**Architecture Decision Records (ADRs)** are an underrated resume signal. A folder with 5–6 short markdown files like `0001-why-rest-over-grpc-internally.md`, `0002-outbox-pattern-for-order-events.md` shows you make deliberate engineering decisions, not just follow tutorials.

---

## 7. CI/CD pipeline (GitHub Actions)

`.github/workflows/ci.yml` — runs on every PR:

```yaml
name: CI
on:
  pull_request:
  push:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - uses: golangci/golangci-lint-action@v6

  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: test
        ports: ['5432:5432']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: go test ./... -race -coverprofile=coverage.out
      - uses: codecov/codecov-action@v4

  build:
    needs: [lint, test]
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [gateway, product, cart, order, payment, inventory, notification]
    steps:
      - uses: actions/checkout@v4
      - uses: docker/build-push-action@v6
        with:
          context: ./services/${{ matrix.service }}
          push: false
          tags: gocommerce/${{ matrix.service }}:${{ github.sha }}
```

`.github/workflows/cd.yml` — runs on merge to `main`, builds + pushes images to GHCR, applies k8s manifests (or just documents the manual `kubectl apply` step if you don't want to pay for a live cluster).

Integration tests should use **testcontainers-go** to spin up real Postgres/Kafka/Redis in CI rather than mocking everything — this is what separates "senior test strategy" from "toy tests."

---

## 8. Observability setup

- **Logging**: `zerolog` or `zap`, structured JSON, correlation ID (trace ID) injected into every log line via middleware.
- **Metrics**: Prometheus client library in each service exposing `/metrics` — track request count, latency histograms, error rate, using `promhttp` middleware on every service.
- **Tracing**: OpenTelemetry SDK, HTTP middleware auto-instruments spans on every incoming and outgoing request, export to Jaeger. This is the single most impressive thing you can show in a demo — pull up the Jaeger UI and show a trace spanning Gateway → Order → Payment → Kafka → Inventory in one waterfall view.
- **Dashboards**: One Grafana dashboard showing request rate, p99 latency, error rate per service — screenshot this for your README.

---

## 9. Local development

`deploy/docker-compose.yml` should bring up: Postgres (one instance, multiple DBs), Redis, Kafka (+ Zookeeper or KRaft mode), Jaeger, Prometheus, Grafana, and all 7 Go services. A single `make dev` should be all it takes to get a fully working system running.

```makefile
dev:
	docker compose -f deploy/docker-compose.yml up --build

generate:
	oapi-codegen -config api/codegen.yaml api/order/openapi.yaml > services/order/api/types.gen.go

test:
	go test ./... -race -cover

lint:
	golangci-lint run ./...

migrate:
	migrate -path services/order/migrations -database $(ORDER_DB_URL) up
```

---

## 10. Testing strategy

| Layer | Approach |
|---|---|
| Domain logic | Pure unit tests, table-driven, no mocks needed (no external deps by design) |
| Application layer | Unit tests with mocked ports (use `mockery` or hand-written fakes) |
| Adapters (Postgres, Kafka) | Integration tests via `testcontainers-go` — real containers, not mocks |
| HTTP handlers | Integration tests using `httptest` for in-process request/response testing |
| End-to-end | A small `docker compose` based e2e suite hitting the gateway and asserting on order flow |

Aim for meaningful coverage on domain + application layers (80%+) rather than chasing 100% everywhere.

---

## 11. Phased roadmap (build order for Claude Code sessions)

Don't ask Claude Code to build everything at once — work through phases, committing after each one.

**Phase 0 — Foundations**
- Repo scaffold, `pkg/` shared libraries (logger, tracing middleware), `oapi-codegen` setup for generating types from OpenAPI specs, Makefile, docker-compose skeleton with just Postgres/Redis/Kafka.

**Phase 1 — Product Service (simplest, proves the pattern)**
- Hexagonal structure, Postgres adapter, REST handler (from the OpenAPI spec), CRUD + search, unit + integration tests, Dockerfile.
- Get this one fully done and tested before moving on — it's your template for every other service.

**Phase 2 — Cart Service**
- Redis adapter, REST handler, TTL logic.

**Phase 3 — API Gateway**
- Routing to Product + Cart services, JWT auth middleware, rate limiting.

**Phase 4 — Order Service + outbox pattern**
- The state machine, outbox table + poller, Kafka publisher.

**Phase 5 — Payment Service**
- Idempotency key handling, mock/Stripe sandbox integration, consumes nothing, gets called synchronously by Order (or via saga — your call on complexity).

**Phase 6 — Inventory + Notification Services**
- Kafka consumers, reservation logic, mock notification sending.

**Phase 7 — Observability**
- Add OpenTelemetry tracing, Prometheus metrics, Jaeger/Grafana to docker-compose, wire into every service.

**Phase 8 — CI/CD + Kubernetes**
- GitHub Actions pipeline, k8s manifests, deploy to a local kind/k3d cluster, record a demo.

**Phase 9 — Resilience polish**
- Circuit breakers, retries, graceful shutdown, chaos-test by killing a service mid-flow and showing the system recovers.

**Phase 10 — Optional: React dashboard**
- Simple live view of orders/inventory, WebSocket or polling against the gateway.

**Phase 11 — Documentation & demo polish**
- ADRs, architecture diagram in the README, a 2-minute Loom-style demo video, load test results (k6 or vegeta) showing throughput numbers — numbers in a README are very persuasive.

---

## 12. How to brief Claude Code per session

For each phase, give Claude Code a focused prompt referencing this doc, e.g.:

> "Using GoCommerce-Project-Documentation.md as the spec, implement Phase 1: Product Service. Follow the hexagonal structure in section 3.1, the data model in section 4, and write table-driven tests for the domain layer. Don't touch other services yet."

This keeps each session scoped and reviewable, which also mirrors how you'd actually work on a real engineering team — another thing worth mentioning in an interview.

---

## 13. Resume bullet points this project supports (once built)

- Designed and built a 7-service Go microservices e-commerce platform using REST APIs (OpenAPI-first) for internal communication and an event-driven architecture with Kafka, implementing the outbox pattern for reliable event delivery.
- Implemented hexagonal architecture across all services, achieving clean separation between domain logic and infrastructure, with 80%+ test coverage using table-driven and container-based integration tests.
- Built full observability stack (Prometheus, Grafana, OpenTelemetry/Jaeger) providing distributed tracing across the entire request lifecycle.
- Designed a CI/CD pipeline in GitHub Actions covering linting, testing with real containerized dependencies, and Docker image builds, deployed to Kubernetes.
- Implemented resilience patterns including circuit breakers, retry-with-backoff, and idempotent payment processing to handle partial failures in a distributed system.
