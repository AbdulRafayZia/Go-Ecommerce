# GoCommerce - Distributed E-Commerce Platform

A production-grade microservices-based e-commerce platform built with Go, demonstrating modern backend engineering practices and patterns.

## Overview

GoCommerce is a distributed e-commerce system showcasing:

- **Microservices Architecture**: 7 independently deployable services
- **Event-Driven Design**: Kafka-based async communication with outbox pattern
- **REST/HTTP Communication**: OpenAPI 3.0 specifications with type-safe code generation
- **Clean Architecture**: Hexagonal architecture in every service
- **Full Observability**: OpenTelemetry tracing, Prometheus metrics, structured logging
- **Resilience Patterns**: Circuit breakers, retries, graceful shutdown, idempotency
- **Modern DevOps**: Docker, Kubernetes, CI/CD with GitHub Actions

## Services

| Service | Port | Description |
|---------|------|-------------|
| **Gateway** | 8080 | REST API gateway, JWT auth, rate limiting |
| **Product** | 50051 | Product catalog and search |
| **Cart** | 50052 | Shopping cart with Redis TTL |
| **Order** | 50053 | Order management with outbox pattern |
| **Payment** | 50054 | Payment processing with idempotency |
| **Inventory** | 50055 | Stock management and reservations |
| **Notification** | 50056 | Event-driven notifications |

## Tech Stack

### Core
- **Go 1.23+**: Modern, performant backend language
- **REST/HTTP + OpenAPI**: Service communication with type-safe code generation
- **PostgreSQL**: Primary data store
- **Redis**: Caching and session management
- **Kafka**: Event streaming (KRaft mode)

### Observability
- **OpenTelemetry + Jaeger**: Distributed tracing
- **Prometheus + Grafana**: Metrics and dashboards
- **Zerolog**: Structured JSON logging

### Development Tools
- **oapi-codegen**: OpenAPI code generation
- **Chi Router**: Lightweight HTTP routing
- **Docker Compose**: Local orchestration
- **Kubernetes**: Container orchestration
- **GitHub Actions**: CI/CD

## Quick Start

### Prerequisites

- Go 1.23 or higher
- Docker and Docker Compose
- Make

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/gocommerce.git
   cd gocommerce
   ```

2. **Install development tools**
   ```bash
   make install-tools
   ```

3. **Download dependencies**
   ```bash
   make deps
   ```

4. **Start infrastructure services**
   ```bash
   make dev
   ```

   This will start:
   - PostgreSQL (port 5432)
   - Redis (port 6379)
   - Kafka (port 9092)
   - Jaeger UI (http://localhost:16686)
   - Prometheus (http://localhost:9090)
   - Grafana (http://localhost:3000)
   - Kafka UI (http://localhost:8080)
   - pgAdmin (http://localhost:5050)

### Development

#### Generate OpenAPI Types

```bash
make api-gen
```

#### Run Tests

```bash
make test                # All tests
make test-coverage       # With coverage report
make test-short          # Quick tests only
```

#### Lint Code

```bash
make lint                # Run linters
make lint-fix            # Auto-fix issues
```

#### Build Services

```bash
make build               # Build all services
make build-service SERVICE=product  # Build specific service
```

## Project Structure

```
gocommerce/
├── services/           # Microservices
│   ├── gateway/
│   ├── product/
│   ├── cart/
│   ├── order/
│   ├── payment/
│   ├── inventory/
│   └── notification/
├── pkg/                # Shared libraries
│   ├── logger/         # Structured logging
│   ├── tracing/        # OpenTelemetry setup
│   ├── middleware/     # HTTP interceptors
│   └── httputil/       # HTTP utilities
├── deploy/             # Deployment configs
│   ├── docker-compose.yml
│   └── k8s/
├── docs/               # Documentation
│   ├── architecture.md
│   └── adr/            # Architecture Decision Records
├── Makefile
└── README.md
```

## Architecture Patterns

### Hexagonal Architecture

Each service follows clean architecture principles:

```
service/
├── cmd/server/main.go           # Composition root
├── internal/
│   ├── domain/                  # Business entities
│   ├── application/             # Use cases
│   ├── ports/                   # Interfaces
│   └── adapters/                # Implementations
│       ├── postgres/
│       ├── kafka/
│       └── http/
├── api/                         # OpenAPI specifications
└── migrations/                  # Database migrations
```

### Outbox Pattern

The Order Service implements the transactional outbox pattern for reliable event publishing:

1. Order and event saved in same transaction
2. Background poller publishes events to Kafka
3. Guarantees at-least-once delivery

### Resilience

- **Circuit Breakers**: Prevent cascading failures
- **Retries with Backoff**: Automatic retry on transient failures
- **Timeouts**: Every HTTP call has a deadline
- **Graceful Shutdown**: Drain in-flight requests before exit
- **Idempotency**: Payment service uses idempotency keys

## Observability

### Distributed Tracing

View end-to-end request traces in Jaeger:
- http://localhost:16686

### Metrics

Prometheus metrics available at:
- http://localhost:9090

Grafana dashboards at:
- http://localhost:3000 (admin/admin)

### Logs

All services output structured JSON logs with:
- Correlation IDs for request tracking
- Trace IDs for distributed tracing
- Contextual fields for debugging

## Database Migrations

Create a new migration:
```bash
make migrate-create SERVICE=order NAME=create_orders_table
```

Run migrations:
```bash
make migrate-up SERVICE=order DB_URL="postgresql://gocommerce:gocommerce_dev_password@localhost:5432/order_db?sslmode=disable"
```

## Testing Strategy

- **Unit Tests**: Domain and application layer (table-driven)
- **Integration Tests**: Using testcontainers-go for real dependencies
- **End-to-End Tests**: Full workflow testing via API gateway

## CI/CD

GitHub Actions pipeline includes:
1. Linting (golangci-lint)
2. Testing (with coverage)
3. Building Docker images
4. Deploying to Kubernetes

## Deployment

### Docker

Build images:
```bash
make docker-build
```

### Kubernetes

Deploy to cluster:
```bash
make k8s-deploy
```

## Development Workflow

1. Create feature branch
2. Implement feature following hexagonal architecture
3. Write tests (unit + integration)
4. Run `make check` (lint + test)
5. Create pull request
6. CI pipeline runs automatically
7. Merge to main triggers deployment

## Contributing

This is a portfolio project, but suggestions and improvements are welcome!

## License

MIT License - feel free to use this project as a learning resource or portfolio template.

## Acknowledgments

Built following best practices from:
- Domain-Driven Design
- Microservices Patterns (Chris Richardson)
- Building Microservices (Sam Newman)
- Production-Ready Microservices (Susan Fowler)

---

**Author**: [Your Name]
**Contact**: [Your Email]
**LinkedIn**: [Your LinkedIn]
