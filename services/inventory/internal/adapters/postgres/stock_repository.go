package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"gocommerce/services/inventory/internal/domain"
	"gocommerce/services/inventory/internal/ports"
)

type StockRepository struct {
	db *sql.DB
}

func NewStockRepository(db *sql.DB) ports.StockRepository {
	return &StockRepository{db: db}
}

func (r *StockRepository) GetByProductID(ctx context.Context, productID string) (*domain.Stock, error) {
	query := `
		SELECT product_id, available, reserved, updated_at
		FROM stock
		WHERE product_id = $1
	`

	var stock domain.Stock
	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&stock.ProductID,
		&stock.Available,
		&stock.Reserved,
		&stock.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrStockNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get stock: %w", err)
	}

	return &stock, nil
}

func (r *StockRepository) Create(ctx context.Context, stock *domain.Stock) error {
	query := `
		INSERT INTO stock (product_id, available, reserved, updated_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(ctx, query,
		stock.ProductID,
		stock.Available,
		stock.Reserved,
		stock.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create stock: %w", err)
	}

	return nil
}

func (r *StockRepository) Update(ctx context.Context, stock *domain.Stock) error {
	query := `
		UPDATE stock
		SET available = $2, reserved = $3, updated_at = $4
		WHERE product_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		stock.ProductID,
		stock.Available,
		stock.Reserved,
		stock.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrStockNotFound
	}

	return nil
}

func (r *StockRepository) CreateOrUpdate(ctx context.Context, stock *domain.Stock) error {
	query := `
		INSERT INTO stock (product_id, available, reserved, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (product_id)
		DO UPDATE SET
			available = EXCLUDED.available,
			reserved = EXCLUDED.reserved,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		stock.ProductID,
		stock.Available,
		stock.Reserved,
		stock.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create or update stock: %w", err)
	}

	return nil
}

func (r *StockRepository) ReserveStock(ctx context.Context, productID string, quantity int) error {
	query := `
		UPDATE stock
		SET available = available - $2,
		    reserved = reserved + $2,
		    updated_at = NOW()
		WHERE product_id = $1
		  AND available >= $2
	`

	result, err := r.db.ExecContext(ctx, query, productID, quantity)
	if err != nil {
		return fmt.Errorf("failed to reserve stock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Either stock doesn't exist or insufficient stock
		_, err := r.GetByProductID(ctx, productID)
		if err == domain.ErrStockNotFound {
			return domain.ErrStockNotFound
		}
		return domain.ErrInsufficientStock
	}

	return nil
}

func (r *StockRepository) ReleaseReservation(ctx context.Context, productID string, quantity int) error {
	query := `
		UPDATE stock
		SET reserved = reserved - $2,
		    available = available + $2,
		    updated_at = NOW()
		WHERE product_id = $1
		  AND reserved >= $2
	`

	result, err := r.db.ExecContext(ctx, query, productID, quantity)
	if err != nil {
		return fmt.Errorf("failed to release reservation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrStockNotFound
	}

	return nil
}

func (r *StockRepository) FulfillReservation(ctx context.Context, productID string, quantity int) error {
	query := `
		UPDATE stock
		SET reserved = reserved - $2,
		    updated_at = NOW()
		WHERE product_id = $1
		  AND reserved >= $2
	`

	result, err := r.db.ExecContext(ctx, query, productID, quantity)
	if err != nil {
		return fmt.Errorf("failed to fulfill reservation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrStockNotFound
	}

	return nil
}

func (r *StockRepository) ListLowStock(ctx context.Context, threshold int) ([]*domain.Stock, error) {
	query := `
		SELECT product_id, available, reserved, updated_at
		FROM stock
		WHERE available < $1
		ORDER BY available ASC
	`

	rows, err := r.db.QueryContext(ctx, query, threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to list low stock: %w", err)
	}
	defer rows.Close()

	var stocks []*domain.Stock
	for rows.Next() {
		var stock domain.Stock
		err := rows.Scan(
			&stock.ProductID,
			&stock.Available,
			&stock.Reserved,
			&stock.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock: %w", err)
		}
		stocks = append(stocks, &stock)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return stocks, nil
}
