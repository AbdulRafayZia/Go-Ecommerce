# ADR 0002: Implement Outbox Pattern for Reliable Event Publishing

**Status**: Accepted

**Date**: 2024-01-20

**Deciders**: Architecture Team

## Context

The Order Service needs to publish events to Kafka when orders are created, updated, or completed. These events trigger downstream workflows in Payment, Inventory, and Notification services.

**The Problem**: Dual-Write Problem

If we write to the database AND publish to Kafka as two separate operations:

```go
// ❌ PROBLEMATIC APPROACH
func (s *OrderService) CreateOrder(order Order) error {
    // Write to database
    err := s.repo.Save(order)
    if err != nil {
        return err
    }

    // Publish event to Kafka
    err = s.publisher.Publish("order.created", order)
    if err != nil {
        // Order saved but event not published!
        // System is now inconsistent
        return err
    }
}
```

**Failure Scenarios:**
1. Database write succeeds, Kafka publish fails → Order exists but no workflow triggered
2. Kafka publish succeeds, database write fails → Event published for non-existent order
3. Service crashes between operations → Inconsistent state

## Decision

We will implement the **Transactional Outbox Pattern** in the Order Service.

## Solution Design

### Database Schema

```sql
CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type TEXT NOT NULL,           -- e.g., "Order"
    aggregate_id UUID NOT NULL,             -- e.g., order_id
    event_type TEXT NOT NULL,               -- e.g., "order.created"
    payload JSONB NOT NULL,                 -- Full event data
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ                -- NULL until published
);

CREATE INDEX idx_outbox_unpublished ON outbox_events(published_at) WHERE published_at IS NULL;
```

### Write Path (Transactional)

```go
func (s *OrderService) CreateOrder(order Order) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 1. Insert order
    err = s.repo.SaveTx(tx, order)
    if err != nil {
        return err
    }

    // 2. Insert event into outbox (same transaction)
    event := OutboxEvent{
        AggregateType: "Order",
        AggregateID:   order.ID,
        EventType:     "order.created",
        Payload:       order.ToJSON(),
    }
    err = s.outbox.InsertTx(tx, event)
    if err != nil {
        return err
    }

    // 3. Commit both writes atomically
    return tx.Commit()
}
```

**Key Guarantee**: Order and event are saved atomically. If either fails, both rollback.

### Publish Path (Asynchronous)

A background **Outbox Poller** runs continuously:

```go
func (p *OutboxPoller) Run(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            p.publishPendingEvents(ctx)
        case <-ctx.Done():
            return
        }
    }
}

func (p *OutboxPoller) publishPendingEvents(ctx context.Context) {
    // 1. Fetch unpublished events (with limit)
    events, err := p.outbox.GetUnpublished(100)
    if err != nil {
        log.Error("failed to fetch events", err)
        return
    }

    for _, event := range events {
        // 2. Publish to Kafka
        err := p.publisher.Publish(event.EventType, event.Payload)
        if err != nil {
            log.Error("failed to publish event", err)
            continue // Retry on next poll
        }

        // 3. Mark as published
        err = p.outbox.MarkPublished(event.ID)
        if err != nil {
            log.Error("failed to mark as published", err)
            // Event will be republished (idempotent consumers required)
        }
    }
}
```

**Delivery Guarantee**: At-least-once delivery. Events may be published multiple times if marking fails.

## Rationale

### Why This Pattern?

**Atomicity**
- Order and event written in single database transaction
- ACID guarantees prevent partial failures

**Reliability**
- Events never lost (persisted in database)
- Poller retries on Kafka failures
- Service crash doesn't lose events

**Separation of Concerns**
- Write path is fast (just DB transaction)
- Publish path is asynchronous (doesn't block user requests)

**Observability**
- Can query outbox table to see pending/failed events
- Easy to debug "event not published" issues

### Alternatives Considered

**1. Synchronous Publish (Dual-Write)**
- ❌ Inconsistency risk as described above
- ❌ Kafka downtime blocks order creation

**2. Two-Phase Commit (2PC)**
- ❌ Requires distributed transaction coordinator
- ❌ High latency and complexity
- ❌ Kafka doesn't support XA transactions

**3. Change Data Capture (CDC)**
- ✅ More elegant solution (e.g., Debezium)
- ❌ Requires additional infrastructure
- ❌ Operational complexity
- 📝 Future enhancement possibility

**4. Event Sourcing**
- ✅ Events as source of truth
- ❌ Full architectural paradigm shift
- ❌ Complex for current team maturity
- 📝 Possible for future v2

## Consequences

### Positive

- **Data Consistency**: Guaranteed by database transaction
- **No Message Loss**: Events persisted durably
- **Debuggability**: Can query outbox table
- **Resilience**: Survives Kafka outages
- **Simple**: Uses existing database, no new infrastructure

### Negative

- **Latency**: Events published asynchronously (1-2s delay typical)
  - *Acceptable*: Order creation is already async workflow
- **At-Least-Once**: Consumers must be idempotent
  - *Mitigation*: Use event IDs for deduplication
- **Outbox Cleanup**: Old events need archival
  - *Mitigation*: Cron job to delete published events older than 7 days
- **Polling Overhead**: Poller queries database frequently
  - *Mitigation*: Index on `published_at`, query only unpublished

### Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Poller crashes | Run multiple poller instances (only one processes via locking) |
| Large event backlog | Horizontal scaling (partition by aggregate_id hash) |
| Kafka downtime | Events accumulate safely in database, auto-recover on Kafka restart |
| Duplicate events | Consumers use idempotency keys from event payload |

## Implementation Checklist

- [x] Create outbox_events table migration
- [ ] Implement outbox repository (insert, fetch, mark published)
- [ ] Modify order creation to use transactional outbox
- [ ] Implement outbox poller service
- [ ] Add metrics (pending events, publish latency, failures)
- [ ] Add alerts (too many pending events, poller stopped)
- [ ] Implement cleanup job for old events
- [ ] Document consumer idempotency requirements

## Monitoring

**Metrics to Track:**
- `outbox_pending_events`: Gauge of unpublished events
- `outbox_publish_duration`: Histogram of publish latency
- `outbox_publish_errors`: Counter of publish failures

**Alerts:**
- Pending events > 1000 for > 5 minutes
- Poller hasn't run in 60 seconds

## Related Decisions

- **ADR 0001**: gRPC for Internal Communication
- **ADR 0003**: Hexagonal Architecture (outbox in adapters layer)

## References

- [Microservices Patterns - Transactional Outbox](https://microservices.io/patterns/data/transactional-outbox.html)
- [Debezium for CDC](https://debezium.io/)
- [Designing Data-Intensive Applications (Martin Kleppmann)](https://dataintensive.net/)
