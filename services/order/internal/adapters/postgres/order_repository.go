package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"gocommerce/services/order/internal/domain"
	"gocommerce/services/order/internal/ports"
)

// OrderRepository is the PostgreSQL implementation of ports.OrderRepository
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository creates a new PostgreSQL order repository
func NewOrderRepository(db *sql.DB) ports.OrderRepository {
	return &OrderRepository{db: db}
}

// Create stores a new order, its items, and domain events in a transaction
func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert order
	orderQuery := `
		INSERT INTO orders (
			id, user_id, status, total_cents, idempotency_key,
			payment_id, tracking_number, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = tx.ExecContext(
		ctx, orderQuery,
		order.ID,
		order.UserID,
		string(order.Status),
		order.TotalCents,
		order.IdempotencyKey,
		order.PaymentID,
		order.TrackingNumber,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert order items
	itemQuery := `
		INSERT INTO order_items (
			order_id, product_id, product_name, quantity, unit_price_cents
		) VALUES ($1, $2, $3, $4, $5)
	`

	for _, item := range order.Items {
		_, err = tx.ExecContext(
			ctx, itemQuery,
			order.ID,
			item.ProductID,
			item.ProductName,
			item.Quantity,
			item.UnitPriceCents,
		)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	// Store domain events in outbox
	if err := r.storeOutboxEvents(ctx, tx, order); err != nil {
		return fmt.Errorf("failed to store outbox events: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Clear events after successful commit
	order.ClearEvents()

	return nil
}

// Update updates an existing order and stores any new domain events
func (r *OrderRepository) Update(ctx context.Context, order *domain.Order) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update order
	query := `
		UPDATE orders
		SET status = $1, payment_id = $2, tracking_number = $3, updated_at = $4
		WHERE id = $5
	`

	result, err := tx.ExecContext(
		ctx, query,
		string(order.Status),
		order.PaymentID,
		order.TrackingNumber,
		order.UpdatedAt,
		order.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrOrderNotFound
	}

	// Store domain events in outbox
	if err := r.storeOutboxEvents(ctx, tx, order); err != nil {
		return fmt.Errorf("failed to store outbox events: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Clear events after successful commit
	order.ClearEvents()

	return nil
}

// GetByID retrieves an order by its ID
func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	// Query order
	orderQuery := `
		SELECT id, user_id, status, total_cents, idempotency_key,
		       payment_id, tracking_number, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	order := &domain.Order{}
	var status string
	var paymentID, trackingNumber sql.NullString

	err := r.db.QueryRowContext(ctx, orderQuery, id).Scan(
		&order.ID,
		&order.UserID,
		&status,
		&order.TotalCents,
		&order.IdempotencyKey,
		&paymentID,
		&trackingNumber,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	order.Status = domain.OrderStatus(status)
	if paymentID.Valid {
		order.PaymentID = &paymentID.String
	}
	if trackingNumber.Valid {
		order.TrackingNumber = &trackingNumber.String
	}

	// Query order items
	itemsQuery := `
		SELECT product_id, product_name, quantity, unit_price_cents
		FROM order_items
		WHERE order_id = $1
	`

	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	items := make([]*domain.OrderItem, 0)
	for rows.Next() {
		item := &domain.OrderItem{}
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.Quantity, &item.UnitPriceCents); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	order.Items = items

	return order, nil
}

// GetByIdempotencyKey retrieves an order by idempotency key
func (r *OrderRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error) {
	// First get the order ID
	var orderID string
	query := "SELECT id FROM orders WHERE idempotency_key = $1"

	err := r.db.QueryRowContext(ctx, query, key).Scan(&orderID)
	if err == sql.ErrNoRows {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order by idempotency key: %w", err)
	}

	// Then get the full order
	return r.GetByID(ctx, orderID)
}

// ExistsByIdempotencyKey checks if an order exists with the given idempotency key
func (r *OrderRepository) ExistsByIdempotencyKey(ctx context.Context, key string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM orders WHERE idempotency_key = $1)"

	err := r.db.QueryRowContext(ctx, query, key).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check order existence: %w", err)
	}

	return exists, nil
}

// List retrieves orders with pagination and filtering
func (r *OrderRepository) List(ctx context.Context, filters ports.ListOrderFilters) ([]*domain.Order, int64, error) {
	// Build WHERE clause
	conditions := make([]string, 0)
	args := make([]interface{}, 0)
	argPos := 1

	if filters.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argPos))
		args = append(args, *filters.UserID)
		argPos++
	}

	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argPos))
		args = append(args, string(*filters.Status))
		argPos++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM orders %s", whereClause)
	var totalCount int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	// Set defaults for pagination
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 || filters.PageSize > 100 {
		filters.PageSize = 20
	}

	offset := (filters.Page - 1) * filters.PageSize

	// Query orders
	query := fmt.Sprintf(`
		SELECT id, user_id, status, total_cents, idempotency_key,
		       payment_id, tracking_number, created_at, updated_at
		FROM orders
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argPos, argPos+1)

	args = append(args, filters.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	orders := make([]*domain.Order, 0)
	for rows.Next() {
		order := &domain.Order{}
		var status string
		var paymentID, trackingNumber sql.NullString

		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&status,
			&order.TotalCents,
			&order.IdempotencyKey,
			&paymentID,
			&trackingNumber,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}

		order.Status = domain.OrderStatus(status)
		if paymentID.Valid {
			order.PaymentID = &paymentID.String
		}
		if trackingNumber.Valid {
			order.TrackingNumber = &trackingNumber.String
		}

		// Note: For list queries, we're not loading items to improve performance
		// If items are needed, use GetByID
		order.Items = make([]*domain.OrderItem, 0)

		orders = append(orders, order)
	}

	return orders, totalCount, nil
}

// storeOutboxEvents stores domain events in the outbox table
func (r *OrderRepository) storeOutboxEvents(ctx context.Context, tx *sql.Tx, order *domain.Order) error {
	events := order.GetEvents()
	if len(events) == 0 {
		return nil
	}

	query := `
		INSERT INTO outbox_events (
			aggregate_type, aggregate_id, event_type, payload, created_at
		) VALUES ($1, $2, $3, $4, $5)
	`

	for _, event := range events {
		// Serialize event to JSON
		payload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		_, err = tx.ExecContext(
			ctx, query,
			"order",
			event.AggregateID(),
			event.EventType(),
			payload,
			event.OccurredAt(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert outbox event: %w", err)
		}
	}

	return nil
}
