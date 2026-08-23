package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"gocommerce/services/payment/internal/domain"
	"gocommerce/services/payment/internal/ports"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) ports.PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	metadataJSON, err := json.Marshal(payment.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO payments (
			id, order_id, amount, currency, status, payment_method,
			provider_payment_id, idempotency_key, failure_reason, metadata,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = r.db.ExecContext(ctx, query,
		payment.ID,
		payment.OrderID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.PaymentMethod,
		nullString(payment.ProviderPaymentID),
		payment.IdempotencyKey,
		nullString(payment.FailureReason),
		metadataJSON,
		payment.CreatedAt,
		payment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	return nil
}

func (r *PaymentRepository) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, amount, currency, status, payment_method,
			   provider_payment_id, idempotency_key, failure_reason, metadata,
			   created_at, updated_at
		FROM payments
		WHERE id = $1
	`

	var payment domain.Payment
	var providerPaymentID, failureReason sql.NullString
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.PaymentMethod,
		&providerPaymentID,
		&payment.IdempotencyKey,
		&failureReason,
		&metadataJSON,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrPaymentNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	payment.ProviderPaymentID = providerPaymentID.String
	payment.FailureReason = failureReason.String

	if err := json.Unmarshal(metadataJSON, &payment.Metadata); err != nil {
		payment.Metadata = make(map[string]interface{})
	}

	return &payment, nil
}

func (r *PaymentRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, amount, currency, status, payment_method,
			   provider_payment_id, idempotency_key, failure_reason, metadata,
			   created_at, updated_at
		FROM payments
		WHERE idempotency_key = $1
	`

	var payment domain.Payment
	var providerPaymentID, failureReason sql.NullString
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.PaymentMethod,
		&providerPaymentID,
		&payment.IdempotencyKey,
		&failureReason,
		&metadataJSON,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrPaymentNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get payment by idempotency key: %w", err)
	}

	payment.ProviderPaymentID = providerPaymentID.String
	payment.FailureReason = failureReason.String

	if err := json.Unmarshal(metadataJSON, &payment.Metadata); err != nil {
		payment.Metadata = make(map[string]interface{})
	}

	return &payment, nil
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, amount, currency, status, payment_method,
			   provider_payment_id, idempotency_key, failure_reason, metadata,
			   created_at, updated_at
		FROM payments
		WHERE order_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var payment domain.Payment
	var providerPaymentID, failureReason sql.NullString
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.PaymentMethod,
		&providerPaymentID,
		&payment.IdempotencyKey,
		&failureReason,
		&metadataJSON,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrPaymentNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get payment by order ID: %w", err)
	}

	payment.ProviderPaymentID = providerPaymentID.String
	payment.FailureReason = failureReason.String

	if err := json.Unmarshal(metadataJSON, &payment.Metadata); err != nil {
		payment.Metadata = make(map[string]interface{})
	}

	return &payment, nil
}

func (r *PaymentRepository) GetByProviderPaymentID(ctx context.Context, providerID string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, amount, currency, status, payment_method,
			   provider_payment_id, idempotency_key, failure_reason, metadata,
			   created_at, updated_at
		FROM payments
		WHERE provider_payment_id = $1
	`

	var payment domain.Payment
	var providerPaymentID, failureReason sql.NullString
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, providerID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.PaymentMethod,
		&providerPaymentID,
		&payment.IdempotencyKey,
		&failureReason,
		&metadataJSON,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrPaymentNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get payment by provider ID: %w", err)
	}

	payment.ProviderPaymentID = providerPaymentID.String
	payment.FailureReason = failureReason.String

	if err := json.Unmarshal(metadataJSON, &payment.Metadata); err != nil {
		payment.Metadata = make(map[string]interface{})
	}

	return &payment, nil
}

func (r *PaymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	metadataJSON, err := json.Marshal(payment.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE payments
		SET status = $1,
		    provider_payment_id = $2,
		    failure_reason = $3,
		    metadata = $4,
		    updated_at = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		payment.Status,
		nullString(payment.ProviderPaymentID),
		nullString(payment.FailureReason),
		metadataJSON,
		payment.UpdatedAt,
		payment.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrPaymentNotFound
	}

	return nil
}

func (r *PaymentRepository) List(ctx context.Context, filter ports.PaymentFilter) ([]*domain.Payment, int, error) {
	// Build query with filters
	query := `
		SELECT id, order_id, amount, currency, status, payment_method,
			   provider_payment_id, idempotency_key, failure_reason, metadata,
			   created_at, updated_at
		FROM payments
		WHERE 1=1
	`

	countQuery := `SELECT COUNT(*) FROM payments WHERE 1=1`

	args := []interface{}{}
	argPos := 1

	// Apply filters
	if filter.OrderID != nil {
		query += fmt.Sprintf(" AND order_id = $%d", argPos)
		countQuery += fmt.Sprintf(" AND order_id = $%d", argPos)
		args = append(args, *filter.OrderID)
		argPos++
	}

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		countQuery += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filter.Status)
		argPos++
	}

	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count payments: %w", err)
	}

	// Add ordering and pagination
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list payments: %w", err)
	}
	defer rows.Close()

	payments := []*domain.Payment{}
	for rows.Next() {
		var payment domain.Payment
		var providerPaymentID, failureReason sql.NullString
		var metadataJSON []byte

		err := rows.Scan(
			&payment.ID,
			&payment.OrderID,
			&payment.Amount,
			&payment.Currency,
			&payment.Status,
			&payment.PaymentMethod,
			&providerPaymentID,
			&payment.IdempotencyKey,
			&failureReason,
			&metadataJSON,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan payment: %w", err)
		}

		payment.ProviderPaymentID = providerPaymentID.String
		payment.FailureReason = failureReason.String

		if err := json.Unmarshal(metadataJSON, &payment.Metadata); err != nil {
			payment.Metadata = make(map[string]interface{})
		}

		payments = append(payments, &payment)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating payments: %w", err)
	}

	return payments, total, nil
}

// Helper function to convert string to sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
