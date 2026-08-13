# ADR 0003: Adopt Hexagonal Architecture for All Services

**Status**: Accepted

**Date**: 2024-01-25

**Deciders**: Architecture Team

## Context

We need a consistent architectural pattern for structuring our microservices that:

1. Separates business logic from infrastructure concerns
2. Makes services testable without external dependencies
3. Allows technology changes without rewriting business logic
4. Provides clear boundaries and dependencies

## Decision

We will adopt **Hexagonal Architecture** (also known as Ports & Adapters) for all services in the GoCommerce platform.

## Pattern Overview

### Layer Structure

```
service/
├── cmd/
│   └── server/
│       └── main.go              # Composition root - wires everything
├── internal/
│   ├── domain/                  # Inner layer - Business logic
│   │   ├── order.go             # Entities
│   │   ├── order_test.go        # Pure unit tests
│   │   └── errors.go            # Domain errors
│   ├── application/             # Middle layer - Use cases
│   │   ├── create_order.go      # Orchestration
│   │   ├── create_order_test.go # Tests with mocked ports
│   │   └── dto.go               # Data transfer objects
│   ├── ports/                   # Interfaces (dependency inversion)
│   │   ├── repository.go        # Storage contract
│   │   ├── publisher.go         # Event publishing contract
│   │   └── payment_client.go    # External service contract
│   └── adapters/                # Outer layer - Implementations
│       ├── http/                # Inbound - REST API handlers
│       │   ├── handler.go
│       │   └── mapper.go        # JSON ↔ Domain
│       ├── postgres/            # Outbound - Data persistence
│       │   ├── repository.go    # Implements ports.Repository
│       │   └── migrations/
│       └── kafka/               # Outbound - Event publishing
│           └── publisher.go     # Implements ports.Publisher
├── api/
│   └── openapi.yaml             # OpenAPI 3.0 specification
└── go.mod
```

### Dependency Rule

**All dependencies point inward:**

```
Adapters → Ports → Application → Domain
   ↑                                ↑
   └── No dependencies ─────────────┘
```

- **Domain**: Zero dependencies (not even on Go standard library where possible)
- **Application**: Depends only on domain + port interfaces
- **Ports**: Depends only on domain types
- **Adapters**: Depends on everything (implements ports, calls application)

## Rationale

### Core Principles

**1. Business Logic Isolation**

Domain layer has no infrastructure dependencies:

```go
// ✅ GOOD - Pure domain logic
package domain

type Order struct {
    ID     string
    Items  []OrderItem
    Total  Money
    Status OrderStatus
}

func (o *Order) AddItem(item OrderItem) error {
    if o.Status != StatusDraft {
        return ErrOrderNotEditable
    }
    o.Items = append(o.Items, item)
    o.Total = o.calculateTotal()
    return nil
}
```

```go
// ❌ BAD - Domain mixed with infrastructure
type Order struct {
    ID     string `gorm:"primaryKey"`  // ← Database concern
    Items  []OrderItem
    Total  Money
}

func (o *Order) Save(db *gorm.DB) error {  // ← Infrastructure leak
    return db.Save(o).Error
}
```

**2. Dependency Inversion**

Application depends on interfaces (ports), not concrete implementations:

```go
// Application layer
type OrderService struct {
    repo      ports.OrderRepository  // Interface, not *PostgresRepo
    publisher ports.EventPublisher   // Interface, not *KafkaPublisher
}

func (s *OrderService) CreateOrder(dto CreateOrderDTO) error {
    order := domain.NewOrder(dto.UserID, dto.Items)

    err := s.repo.Save(order)  // Don't care if it's Postgres, MongoDB, etc.
    if err != nil {
        return err
    }

    err = s.publisher.Publish("order.created", order)  // Don't care if Kafka, RabbitMQ
    return err
}
```

**3. Testability**

Domain and application layers can be tested without infrastructure:

```go
// Domain tests - no mocks needed
func TestOrder_AddItem(t *testing.T) {
    order := domain.NewOrder("user-123")
    item := domain.OrderItem{ProductID: "prod-1", Quantity: 2}

    err := order.AddItem(item)

    assert.NoError(t, err)
    assert.Len(t, order.Items, 1)
}

// Application tests - mock only ports
func TestOrderService_CreateOrder(t *testing.T) {
    mockRepo := &mocks.OrderRepository{}
    mockPublisher := &mocks.EventPublisher{}
    service := NewOrderService(mockRepo, mockPublisher)

    dto := CreateOrderDTO{UserID: "user-123", Items: []...}

    err := service.CreateOrder(dto)

    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

**4. Technology Swappability**

Can replace infrastructure without touching business logic:

- Postgres → MongoDB: Rewrite adapter, ports/application unchanged
- Kafka → RabbitMQ: Rewrite adapter, ports/application unchanged
- REST → GraphQL: Add new adapter, reuse same application layer

### Why Not Layered Architecture?

**Layered (N-Tier) Architecture:**

```
API Layer → Service Layer → Repository Layer → Database
```

**Problems:**
- Service layer often depends on repository interfaces (OK)
- But domain models often have database annotations (NOT OK)
- Hard to test service without database
- Technology leaks into business logic

**Hexagonal Advantage:**
- Clear separation: business logic never sees database/HTTP/messaging
- Tests run in milliseconds (no I/O)
- New features touch only necessary layers

## Implementation Guidelines

### 1. Domain Layer

**What belongs here:**
- Entities and value objects
- Business rules and invariants
- Domain events (not infrastructure events)
- Domain errors

**Rules:**
- No imports from `adapters/`, `ports/`, or frameworks
- Only pure Go (and maybe standard library)
- Methods return domain types or domain errors

### 2. Application Layer

**What belongs here:**
- Use cases (CreateOrder, UpdateInventory, etc.)
- Orchestration of domain objects
- Transaction boundaries
- DTOs for input/output

**Rules:**
- Depends on domain + port interfaces
- No direct infrastructure calls
- Handles cross-cutting concerns (transactions, events)

### 3. Ports Layer

**What belongs here:**
- Repository interfaces
- External service interfaces
- Event publisher interfaces

**Rules:**
- Methods accept/return domain types
- No implementation details (e.g., no SQL queries)

Example:
```go
type OrderRepository interface {
    Save(order *domain.Order) error
    FindByID(id string) (*domain.Order, error)
    FindByUserID(userID string) ([]*domain.Order, error)
}
```

### 4. Adapters Layer

**What belongs here:**
- REST/HTTP handlers
- Database repositories
- Kafka publishers
- External API clients

**Rules:**
- Implements port interfaces
- Handles infrastructure concerns (serialization, connection pooling)
- Maps between domain types and external formats

Example:
```go
// Inbound adapter (HTTP handler)
func (h *OrderHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
    // 1. Map JSON request → DTO
    dto := mapProtoToDTO(req)

    // 2. Call application layer
    order, err := h.orderService.CreateOrder(dto)
    if err != nil {
        return nil, httputil.ToHTTPStatus(err)
    }

    // 3. Map domain → JSON response
    return mapDomainToProto(order), nil
}

// Outbound adapter (Postgres repository)
type PostgresOrderRepository struct {
    db *sql.DB
}

func (r *PostgresOrderRepository) Save(order *domain.Order) error {
    // Map domain.Order → database row
    row := mapDomainToRow(order)

    _, err := r.db.Exec("INSERT INTO orders (...) VALUES (...)", row...)
    return err
}
```

### 5. Composition Root (main.go)

Wires everything together:

```go
func main() {
    // Infrastructure
    db := setupDatabase()
    kafkaProducer := setupKafka()

    // Adapters (outbound)
    orderRepo := postgres.NewOrderRepository(db)
    eventPublisher := kafka.NewEventPublisher(kafkaProducer)

    // Application
    orderService := application.NewOrderService(orderRepo, eventPublisher)

    // Adapters (inbound)
    httpHandler := http.NewOrderHandler(orderService)

    // Start server
    server := setupHTTPServer(httpHandler)
    server.Serve()
}
```

## Consequences

### Positive

✅ **Testability**: Domain/application tests run without database/Kafka
✅ **Maintainability**: Clear boundaries, easy to locate code
✅ **Flexibility**: Swap implementations without changing business logic
✅ **Onboarding**: New developers understand structure quickly
✅ **Interview Signal**: Shows senior-level architectural thinking

### Negative

❌ **Boilerplate**: More files and interfaces than simple layered architecture
- *Mitigation*: Code generation for CRUD operations

❌ **Learning Curve**: Junior developers may struggle initially
- *Mitigation*: Detailed documentation, pair programming

❌ **Mapping Overhead**: DTO → Domain → Proto conversions
- *Mitigation*: Clear naming conventions, mapper utilities

### Neutral

⚖️ **More Abstraction**: Some see this as over-engineering for simple CRUD
- *Counterpoint*: Pays off as complexity grows, prevents "big ball of mud"

## Validation

**Success Criteria:**
1. Can run all domain tests without `docker-compose up`
2. Can run all application tests with only mocked ports
3. Integration tests use real infrastructure via testcontainers
4. New feature touches only 2-3 layers (domain, application, one adapter)

**Example Test Run:**
```bash
# Fast (< 1s) - no infrastructure
go test ./internal/domain/...
go test ./internal/application/...

# Medium (2-5s) - mocked infrastructure
go test ./internal/adapters/http/...

# Slow (10-30s) - real infrastructure
go test -tags=integration ./...
```

## Related Decisions

- **ADR 0001**: REST for Service Communication (affects adapter layer)
- **ADR 0002**: Outbox Pattern (implemented in outbound adapter)

## References

- [Hexagonal Architecture (Alistair Cockburn)](https://alistair.cockburn.us/hexagonal-architecture/)
- [Clean Architecture (Robert C. Martin)](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Domain-Driven Design (Eric Evans)](https://www.domainlanguage.com/ddd/)
- [Implementing Domain-Driven Design (Vaughn Vernon)](https://vaughnvernon.com/books/)
