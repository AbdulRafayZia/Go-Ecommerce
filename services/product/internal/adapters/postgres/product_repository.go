package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"gocommerce/services/product/internal/domain"
	"gocommerce/services/product/internal/ports"
)

// ProductRepository is the PostgreSQL implementation of ports.ProductRepository
type ProductRepository struct {
	db *sql.DB
}

// NewProductRepository creates a new PostgreSQL product repository
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create stores a new product
func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	query := `
		INSERT INTO products (
			id, name, description, price_cents, currency, category_id,
			stock, active, image_url, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(
		ctx, query,
		product.ID,
		product.Name,
		nullString(product.Description),
		product.PriceCents,
		product.Currency,
		product.CategoryID,
		product.Stock,
		product.Active,
		product.ImageURL,
		product.CreatedAt,
		product.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert product: %w", err)
	}

	return nil
}

// GetByID retrieves a product by its ID
func (r *ProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	query := `
		SELECT id, name, description, price_cents, currency, category_id,
		       stock, active, image_url, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	product := &domain.Product{}
	var description, categoryID, imageURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&description,
		&product.PriceCents,
		&product.Currency,
		&categoryID,
		&product.Stock,
		&product.Active,
		&imageURL,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Map nullable fields
	product.Description = description.String
	if categoryID.Valid {
		product.CategoryID = &categoryID.String
	}
	if imageURL.Valid {
		product.ImageURL = &imageURL.String
	}

	return product, nil
}

// Update updates an existing product
func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	query := `
		UPDATE products
		SET name = $2, description = $3, price_cents = $4, currency = $5,
		    category_id = $6, stock = $7, active = $8, image_url = $9, updated_at = $10
		WHERE id = $1
	`

	result, err := r.db.ExecContext(
		ctx, query,
		product.ID,
		product.Name,
		nullString(product.Description),
		product.PriceCents,
		product.Currency,
		product.CategoryID,
		product.Stock,
		product.Active,
		product.ImageURL,
		product.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrProductNotFound
	}

	return nil
}

// Delete soft deletes a product (sets active = false)
func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE products SET active = false, updated_at = NOW() WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrProductNotFound
	}

	return nil
}

// List retrieves products with pagination and filtering
func (r *ProductRepository) List(ctx context.Context, filters ports.ListProductFilters) ([]*domain.Product, int64, error) {
	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argPos := 1

	if filters.ActiveOnly {
		conditions = append(conditions, fmt.Sprintf("active = $%d", argPos))
		args = append(args, true)
		argPos++
	}

	if filters.CategoryID != nil && *filters.CategoryID != "" {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argPos))
		args = append(args, *filters.CategoryID)
		argPos++
	}

	if filters.Search != nil && *filters.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argPos, argPos))
		searchTerm := "%" + *filters.Search + "%"
		args = append(args, searchTerm)
		argPos++
	}

	if filters.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price_cents >= $%d", argPos))
		args = append(args, *filters.MinPrice)
		argPos++
	}

	if filters.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price_cents <= $%d", argPos))
		args = append(args, *filters.MaxPrice)
		argPos++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total items
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products %s", whereClause)
	var totalItems int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalItems)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Query products with pagination
	query := fmt.Sprintf(`
		SELECT id, name, description, price_cents, currency, category_id,
		       stock, active, image_url, created_at, updated_at
		FROM products
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argPos, argPos+1)

	limit := filters.PageSize
	offset := (filters.Page - 1) * filters.PageSize
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		product := &domain.Product{}
		var description, categoryID, imageURL sql.NullString

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&description,
			&product.PriceCents,
			&product.Currency,
			&categoryID,
			&product.Stock,
			&product.Active,
			&imageURL,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}

		// Map nullable fields
		product.Description = description.String
		if categoryID.Valid {
			product.CategoryID = &categoryID.String
		}
		if imageURL.Valid {
			product.ImageURL = &imageURL.String
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return products, totalItems, nil
}

// ExistsByName checks if a product with the given name exists
func (r *ProductRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE name = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check product existence: %w", err)
	}

	return exists, nil
}

// nullString converts a string to sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
