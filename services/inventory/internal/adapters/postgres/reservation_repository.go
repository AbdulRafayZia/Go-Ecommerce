package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gocommerce/services/inventory/internal/domain"
	"gocommerce/services/inventory/internal/ports"
)

type ReservationRepository struct {
	db *sql.DB
}

func NewReservationRepository(db *sql.DB) ports.ReservationRepository {
	return &ReservationRepository{db: db}
}

func (r *ReservationRepository) Create(ctx context.Context, reservation *domain.Reservation) error {
	query := `
		INSERT INTO reservations (id, order_id, product_id, quantity, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		reservation.ID,
		reservation.OrderID,
		reservation.ProductID,
		reservation.Quantity,
		reservation.Status,
		reservation.CreatedAt,
		reservation.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create reservation: %w", err)
	}

	return nil
}

func (r *ReservationRepository) GetByID(ctx context.Context, id string) (*domain.Reservation, error) {
	query := `
		SELECT id, order_id, product_id, quantity, status, created_at, updated_at
		FROM reservations
		WHERE id = $1
	`

	var reservation domain.Reservation
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&reservation.ID,
		&reservation.OrderID,
		&reservation.ProductID,
		&reservation.Quantity,
		&reservation.Status,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrReservationNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	return &reservation, nil
}

func (r *ReservationRepository) GetByOrderID(ctx context.Context, orderID string) ([]*domain.Reservation, error) {
	query := `
		SELECT id, order_id, product_id, quantity, status, created_at, updated_at
		FROM reservations
		WHERE order_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reservations by order: %w", err)
	}
	defer rows.Close()

	var reservations []*domain.Reservation
	for rows.Next() {
		var reservation domain.Reservation
		err := rows.Scan(
			&reservation.ID,
			&reservation.OrderID,
			&reservation.ProductID,
			&reservation.Quantity,
			&reservation.Status,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}
		reservations = append(reservations, &reservation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return reservations, nil
}

func (r *ReservationRepository) GetByOrderAndProduct(ctx context.Context, orderID, productID string) (*domain.Reservation, error) {
	query := `
		SELECT id, order_id, product_id, quantity, status, created_at, updated_at
		FROM reservations
		WHERE order_id = $1 AND product_id = $2
	`

	var reservation domain.Reservation
	err := r.db.QueryRowContext(ctx, query, orderID, productID).Scan(
		&reservation.ID,
		&reservation.OrderID,
		&reservation.ProductID,
		&reservation.Quantity,
		&reservation.Status,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrReservationNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	return &reservation, nil
}

func (r *ReservationRepository) Update(ctx context.Context, reservation *domain.Reservation) error {
	query := `
		UPDATE reservations
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		reservation.ID,
		reservation.Status,
		reservation.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update reservation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrReservationNotFound
	}

	return nil
}

func (r *ReservationRepository) ListPendingReservations(ctx context.Context, limit int) ([]*domain.Reservation, error) {
	query := `
		SELECT id, order_id, product_id, quantity, status, created_at, updated_at
		FROM reservations
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, domain.ReservationStatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending reservations: %w", err)
	}
	defer rows.Close()

	var reservations []*domain.Reservation
	for rows.Next() {
		var reservation domain.Reservation
		err := rows.Scan(
			&reservation.ID,
			&reservation.OrderID,
			&reservation.ProductID,
			&reservation.Quantity,
			&reservation.Status,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}
		reservations = append(reservations, &reservation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return reservations, nil
}

func (r *ReservationRepository) ListExpiredReservations(ctx context.Context, olderThan time.Duration) ([]*domain.Reservation, error) {
	query := `
		SELECT id, order_id, product_id, quantity, status, created_at, updated_at
		FROM reservations
		WHERE status = $1
		  AND created_at < $2
		ORDER BY created_at ASC
	`

	cutoffTime := time.Now().Add(-olderThan)

	rows, err := r.db.QueryContext(ctx, query, domain.ReservationStatusPending, cutoffTime)
	if err != nil {
		return nil, fmt.Errorf("failed to list expired reservations: %w", err)
	}
	defer rows.Close()

	var reservations []*domain.Reservation
	for rows.Next() {
		var reservation domain.Reservation
		err := rows.Scan(
			&reservation.ID,
			&reservation.OrderID,
			&reservation.ProductID,
			&reservation.Quantity,
			&reservation.Status,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}
		reservations = append(reservations, &reservation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return reservations, nil
}

func (r *ReservationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM reservations WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete reservation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrReservationNotFound
	}

	return nil
}
