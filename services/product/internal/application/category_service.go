package application

import (
	"context"
	"fmt"

	"gocommerce/services/product/internal/domain"
	"gocommerce/services/product/internal/ports"

	"github.com/google/uuid"
)

// CategoryService orchestrates category-related use cases
type CategoryService struct {
	categoryRepo ports.CategoryRepository
}

// NewCategoryService creates a new category service
func NewCategoryService(categoryRepo ports.CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

// CreateCategoryDTO represents the input for creating a category
type CreateCategoryDTO struct {
	Name        string
	Description *string
	ParentID    *string
}

// CreateCategory creates a new category
func (s *CategoryService) CreateCategory(ctx context.Context, dto CreateCategoryDTO) (*domain.Category, error) {
	// Validate parent category if provided
	if dto.ParentID != nil && *dto.ParentID != "" {
		_, err := s.categoryRepo.GetByID(ctx, *dto.ParentID)
		if err != nil {
			return nil, fmt.Errorf("invalid parent category: %w", err)
		}
	}

	// Check if category name already exists
	exists, err := s.categoryRepo.ExistsByName(ctx, dto.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check category existence: %w", err)
	}
	if exists {
		return nil, domain.ErrCategoryAlreadyExists
	}

	// Create category domain entity
	category, err := domain.NewCategory(dto.Name, dto.Description, dto.ParentID)
	if err != nil {
		return nil, err
	}

	category.ID = uuid.New().String()

	// Persist category
	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

// GetCategory retrieves a category by ID
func (s *CategoryService) GetCategory(ctx context.Context, id string) (*domain.Category, error) {
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return category, nil
}

// ListCategories retrieves all categories
func (s *CategoryService) ListCategories(ctx context.Context) ([]*domain.Category, error) {
	categories, err := s.categoryRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	return categories, nil
}

// GetSubcategories retrieves all subcategories of a parent category
func (s *CategoryService) GetSubcategories(ctx context.Context, parentID string) ([]*domain.Category, error) {
	// Validate parent exists
	_, err := s.categoryRepo.GetByID(ctx, parentID)
	if err != nil {
		return nil, err
	}

	subcategories, err := s.categoryRepo.GetByParentID(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subcategories: %w", err)
	}

	return subcategories, nil
}
