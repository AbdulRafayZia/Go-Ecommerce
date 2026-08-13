# GoCommerce Architecture

## System Architecture

GoCommerce follows a microservices architecture pattern with event-driven communication for async workflows.

## High-Level Architecture Diagram

```
                                    ┌─────────────────┐
                                    │   React/Web     │
                                    │   Frontend      │
                                    └────────┬────────┘
                                             │ REST/HTTP
                                             │
                                    ┌────────▼────────┐
                                    │  API Gateway    │
                                    │  (Port 8080)    │
                                    │  - Auth (JWT)   │
                                    │  - Rate Limit   │
                                    └────────┬────────┘
                                             │ REST/HTTP
                     ┌───────────────────────┼───────────────────────┐
                     │                       │                       │
            ┌────────▼────────┐    ┌────────▼────────┐    ┌────────▼────────┐
            │  Product Svc    │    │   Cart Svc      │    │   Order Svc     │
            │  (Port 50051)   │    │  (Port 50052)   │    │  (Port 50053)   │
            │  - Catalog      │    │  - Sessions     │    │  - State Machine│
            │  - Search       │    │  - Redis TTL    │    │  - Outbox       │
            └────────┬────────┘    └────────┬────────┘    └────────┬────────┘
                     │                       │                       │
                     │              ┌────────▼────────┐              │
                     │              │    Redis        │              │
                     │              │  (Port 6379)    │              │
                     │              └─────────────────┘              │
                     │                                               │
                     │              ┌────────▼────────┐              │
                     └──────────────┤   PostgreSQL    ├──────────────┘
                                    │  (Port 5432)    │
                                    │  - product_db   │
                                    │  - order_db     │
                                    │  - payment_db   │
                                    │  - inventory_db │
                                    └────────┬────────┘
                                             │
                                    ┌────────▼────────┐
                                    │      Kafka      │
                                    │  (Port 9092)    │
                                    │  Events:        │
                                    │  - order.created│
                                    │  - payment.done │
                                    │  - inventory.*  │
                                    └────────┬────────┘
                                             │
                     ┌───────────────────────┼───────────────────────┐
            ┌────────▼────────┐    ┌────────▼────────┐    ┌────────▼────────┐
            │  Payment Svc    │    │ Inventory Svc   │    │ Notification    │
            │  (Port 50054)   │    │  (Port 50055)   │    │     Svc         │
            │  - Idempotency  │    │  - Reservations │    │  (Port 50056)   │
            │  - Stripe Mock  │    │  - Stock Levels │    │  - Email Mock   │
            └─────────────────┘    └─────────────────┘    └─────────────────┘

                                    Observability Layer
                     ┌──────────────────────────────────────────────┐
                     │  Jaeger (16686)  │ Prometheus (9090) │       │
                     │  - Tracing       │ - Metrics         │       │
                     │                  │                   │       │
                     │  Grafana (3000)                              │
                     │  - Dashboards                                │
                     └──────────────────────────────────────────────┘
```

## Communication Patterns

### Synchronous (REST/HTTP)

- Gateway → All Services
- Order → Payment (for payment initiation)
- Services expose health checks

**Why REST/HTTP?**
- Strong typing via OpenAPI specifications
- Simple and well-understood by all developers
- Easy to debug (JSON is human-readable)
- Code generation for type safety (oapi-codegen)

### Asynchronous (Kafka)

**Topics:**
- `order.created` - Published by Order Service
- `payment.completed` - Published by Payment Service
- `payment.failed` - Published by Payment Service
- `inventory.reserved` - Published by Inventory Service
- `order.shipped` - Published by Order Service

**Consumers:**
- Payment Service → `order.created`
- Inventory Service → `payment.completed`
- Notification Service → All events

**Why Kafka?**
- Durable event log
- Decouples services
- Enables event sourcing patterns
- Scales horizontally

## Data Architecture

### Database Per Service Pattern

Each service owns its data:

```
Product Service  → product_db (PostgreSQL)
Order Service    → order_db (PostgreSQL)
Payment Service  → payment_db (PostgreSQL)
Inventory Service→ inventory_db (PostgreSQL)
Cart Service     → Redis (TTL-based)
```

**Benefits:**
- Service independence
- Schema evolution per service
- Technology flexibility
- Failure isolation

**Tradeoffs:**
- No foreign keys across services
- Eventual consistency required
- Distributed transactions complexity

### Data Consistency

**Strong Consistency:**
- Within a service (ACID transactions)
- Order creation + outbox event

**Eventual Consistency:**
- Across services (via events)
- Order created → Payment processed → Inventory reserved

## Hexagonal Architecture (Per Service)

```
┌─────────────────────────────────────────────────────────────┐
│                         Service                             │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                    Domain Layer                        │  │
│  │  - Pure business logic                                 │  │
│  │  - No framework dependencies                           │  │
│  │  - Entities, Value Objects, Domain Errors              │  │
│  └───────────────────┬───────────────────────────────────┘  │
│                      │                                       │
│  ┌───────────────────▼───────────────────────────────────┐  │
│  │                Application Layer                       │  │
│  │  - Use cases / services                                │  │
│  │  - Orchestrates domain + ports                         │  │
│  │  - Transaction boundaries                              │  │
│  └───────────────────┬───────────────────────────────────┘  │
│                      │                                       │
│  ┌───────────────────▼───────────────────────────────────┐  │
│  │                   Ports Layer                          │  │
│  │  - Repository interfaces                               │  │
│  │  - Publisher interfaces                                │  │
│  │  - External service interfaces                         │  │
│  └───────────────────┬───────────────────────────────────┘  │
│                      │                                       │
│  ┌───────────────────▼───────────────────────────────────┐  │
│  │                Adapters Layer                          │  │
│  │  Inbound:          │         Outbound:                 │  │
│  │  - REST/HTTP handlers   │         - Postgres repos          │  │
│  │  - REST handlers   │         - Kafka publishers        │  │
│  │                    │         - External APIs           │  │
│  └────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**Benefits:**
- Testable business logic
- Swappable infrastructure
- Clear dependency direction (inward)
- Framework independence

## Resilience Patterns

### 1. Circuit Breaker
- Prevents cascade failures
- Used: Gateway → Services, Order → Payment
- Library: `sony/gobreaker`

### 2. Retry with Backoff
- Handles transient failures
- Exponential backoff
- Library: `cenkalti/backoff`

### 3. Timeouts
- Every REST/HTTP call has deadline
- Context propagation

### 4. Graceful Shutdown
- SIGTERM handling
- Drain in-flight requests
- Close connections cleanly

### 5. Idempotency
- Payment Service tracks idempotency keys
- Prevents duplicate charges
- Client-supplied or generated

### 6. Rate Limiting
- Token bucket algorithm
- Per-IP or per-user
- Redis-backed counters

## Security

### Authentication & Authorization
- JWT tokens issued by Gateway
- Token validation on every request
- Claims include user_id, roles

### Service-to-Service
- Currently: Internal network trust
- Future: mTLS between services

### Data Protection
- Secrets in environment variables
- No credentials in code/configs
- Database encryption at rest (production)

## Observability

### Tracing (OpenTelemetry + Jaeger)
- Trace ID injected at Gateway
- Propagated via REST/HTTP metadata
- Spans for every operation
- Visualize full request path

### Metrics (Prometheus)
- Request count, latency, errors
- REST/HTTP method-level metrics
- Custom business metrics
- Scraped every 15s

### Logging (Zerolog)
- Structured JSON logs
- Correlation ID in every log
- Log levels: DEBUG, INFO, WARN, ERROR
- Centralized in production (future: ELK stack)

## Deployment

### Local Development
- Docker Compose
- All services + infrastructure
- Single command: `make dev`

### Production
- Kubernetes cluster
- Helm charts (optional)
- Horizontal Pod Autoscaling
- Managed PostgreSQL, Redis, Kafka

## Scalability

### Horizontal Scaling
- All services are stateless
- Cart data in shared Redis
- Scale independently per load

### Database Scaling
- Read replicas for heavy read services (Product)
- Connection pooling
- Prepared statements

### Kafka Scaling
- Partitioned topics
- Consumer groups
- Parallel processing

## Future Enhancements

1. **API Versioning**: v2 endpoints for breaking changes
2. **GraphQL Gateway**: Alternative to REST
3. **Service Mesh**: Istio for mTLS, traffic management
4. **CQRS**: Separate read/write models for Order Service
5. **Event Sourcing**: Full event store for audit trail
6. **Saga Pattern**: Distributed transaction coordination
7. **CDN**: Static asset delivery
8. **Multi-region**: Geographic distribution
