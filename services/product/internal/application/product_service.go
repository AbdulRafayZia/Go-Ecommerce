package application

import (
	"context"
	"fmt"

	"gocommerce/services/product/internal/domain"
	"gocommerce/services/product/internal/ports"

	"github.com/google/uuid"
)

// ProductService orchestrates product-related use cases
// This is the application layer that coordinates domain logic and infrastructure
type ProductService struct {
	productRepo  ports.ProductRepository
	categoryRepo ports.CategoryRepository
}

// NewProductService creates a new product service
func NewProductService(productRepo ports.ProductRepository, categoryRepo ports.CategoryRepository) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

// CreateProductDTO represents the input for creating a product
type CreateProductDTO struct {
	Name        string
	Description string
	PriceCents  int
	Currency    string
	CategoryID  *string
	Stock       int
	ImageURL    *string
}

// UpdateProductDTO represents the input for updating a product
type UpdateProductDTO struct {
	Name        *string
	Description *string
	PriceCents  *int
	Currency    *string
	CategoryID  *string
	Stock       *int
	Active      *bool
	ImageURL    *string
}

// ListProductsDTO represents the input for listing products
type ListProductsDTO struct {
	Page       int
	PageSize   int
	CategoryID *string
	Search     *string
	MinPrice   *int
	MaxPrice   *int
}

// ProductListResult represents the result of listing products
type ProductListResult struct {
	Products   []*domain.Product
	TotalItems int64
	TotalPages int
	Page       int
	PageSize   int
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(ctx context.Context, dto CreateProductDTO) (*domain.Product, error) {
	// Validate category if provided
	if dto.CategoryID != nil && *dto.CategoryID != "" {
		_, err := s.categoryRepo.GetByID(ctx, *dto.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("invalid category: %w", err)
		}
	}

	// Check if product name already exists
	exists, err := s.productRepo.ExistsByName(ctx, dto.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check product existence: %w", err)
	}
	if exists {
		return nil, domain.ErrProductAlreadyExists
	}

	// Create product domain entity
	product, err := domain.NewProduct(dto.Name, dto.Description, dto.PriceCents, dto.Currency, dto.Stock)
	if err != nil {
		return nil, err
	}

	// Set optional fields
	product.ID = uuid.New().String()
	if dto.CategoryID != nil {
		product.SetCategory(dto.CategoryID)
	}
	if dto.ImageURL != nil {
		product.SetImageURL(dto.ImageURL)
	}

	// Persist product
	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return product, nil
}

// UpdateProduct updates an existing product
func (s *ProductService) UpdateProduct(ctx context.Context, id string, dto UpdateProductDTO) (*domain.Product, error) {
	// Retrieve existing product
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if dto.Name != nil {
		if err := product.UpdateName(*dto.Name); err != nil {
			return nil, err
		}
	}

	if dto.Description != nil {
		product.UpdateDescription(*dto.Description)
	}

	if dto.PriceCents != nil {
		if err := product.UpdatePrice(*dto.PriceCents); err != nil {
			return nil, err
		}
	}

	if dto.Currency != nil {
		product.Currency = *dto.Currency
	}

	if dto.Stock != nil {
		if err := product.UpdateStock(*dto.Stock); err != nil {
			return nil, err
		}
	}

	if dto.Active != nil {
		if *dto.Active {
			product.Activate()
		} else {
			product.Deactivate()
		}
	}

	if dto.CategoryID != nil {
		// Validate category if changing
		if *dto.CategoryID != "" {
			_, err := s.categoryRepo.GetByID(ctx, *dto.CategoryID)
			if err != nil {
				return nil, fmt.Errorf("invalid category: %w", err)
			}
		}
		product.SetCategory(dto.CategoryID)
	}

	if dto.ImageURL != nil {
		product.SetImageURL(dto.ImageURL)
	}

	// Persist changes
	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

// DeleteProduct soft deletes a product
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	// Check if product exists
	_, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Soft delete
	if err := s.productRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// ListProducts retrieves products with pagination and filtering
func (s *ProductService) ListProducts(ctx context.Context, dto ListProductsDTO) (*ProductListResult, error) {
	// Set defaults
	if dto.Page < 1 {
		dto.Page = 1
	}
	if dto.PageSize < 1 || dto.PageSize > 100 {
		dto.PageSize = 20
	}

	// Build filters
	filters := ports.ListProductFilters{
		Page:       dto.Page,
		PageSize:   dto.PageSize,
		CategoryID: dto.CategoryID,
		Search:     dto.Search,
		MinPrice:   dto.MinPrice,
		MaxPrice:   dto.MaxPrice,
		ActiveOnly: true, // Only show active products by default
	}

	// Retrieve products
	products, totalItems, err := s.productRepo.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Calculate total pages
	totalPages := int(totalItems) / dto.PageSize
	if int(totalItems)%dto.PageSize > 0 {
		totalPages++
	}

	return &ProductListResult{
		Products:   products,
		TotalItems: totalItems,
		TotalPages: totalPages,
		Page:       dto.Page,
		PageSize:   dto.PageSize,
	}, nil
}
