package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"gocommerce/services/product/internal/domain"
)

// CategoryRepository is the PostgreSQL implementation of ports.CategoryRepository
type CategoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository creates a new PostgreSQL category repository
func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Create stores a new category
func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	query := `
		INSERT INTO categories (id, name, description, parent_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(
		ctx, query,
		category.ID,
		category.Name,
		nullString(ptrToString(category.Description)),
		nullString(ptrToString(category.ParentID)),
		category.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert category: %w", err)
	}

	return nil
}

// GetByID retrieves a category by its ID
func (r *CategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	query := `
		SELECT id, name, description, parent_id, created_at
		FROM categories
		WHERE id = $1
	`

	category := &domain.Category{}
	var description, parentID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&description,
		&parentID,
		&category.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrCategoryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	// Map nullable fields
	if description.Valid {
		category.Description = &description.String
	}
	if parentID.Valid {
		category.ParentID = &parentID.String
	}

	return category, nil
}

// Update updates an existing category
func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	query := `
		UPDATE categories
		SET name = $2, description = $3, parent_id = $4
		WHERE id = $1
	`

	result, err := r.db.ExecContext(
		ctx, query,
		category.ID,
		category.Name,
		nullString(ptrToString(category.Description)),
		nullString(ptrToString(category.ParentID)),
	)

	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrCategoryNotFound
	}

	return nil
}

// Delete deletes a category
func (r *CategoryRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM categories WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrCategoryNotFound
	}

	return nil
}

// List retrieves all categories
func (r *CategoryRepository) List(ctx context.Context) ([]*domain.Category, error) {
	query := `
		SELECT id, name, description, parent_id, created_at
		FROM categories
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		category := &domain.Category{}
		var description, parentID sql.NullString

		err := rows.Scan(
			&category.ID,
			&category.Name,
			&description,
			&parentID,
			&category.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}

		// Map nullable fields
		if description.Valid {
			category.Description = &description.String
		}
		if parentID.Valid {
			category.ParentID = &parentID.String
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return categories, nil
}

// ExistsByName checks if a category with the given name exists
func (r *CategoryRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check category existence: %w", err)
	}

	return exists, nil
}

// GetByParentID retrieves all categories with the given parent ID
func (r *CategoryRepository) GetByParentID(ctx context.Context, parentID string) ([]*domain.Category, error) {
	query := `
		SELECT id, name, description, parent_id, created_at
		FROM categories
		WHERE parent_id = $1
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories by parent: %w", err)
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		category := &domain.Category{}
		var description, parentID sql.NullString

		err := rows.Scan(
			&category.ID,
			&category.Name,
			&description,
			&parentID,
			&category.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}

		// Map nullable fields
		if description.Valid {
			category.Description = &description.String
		}
		if parentID.Valid {
			category.ParentID = &parentID.String
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return categories, nil
}

// ptrToString converts a string pointer to string (empty if nil)
func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
