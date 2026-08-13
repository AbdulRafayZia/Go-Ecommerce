package mocks

import (
	"context"
	"sync"

	"gocommerce/services/product/internal/domain"
	"gocommerce/services/product/internal/ports"
)

// MockProductRepository is a mock implementation of ports.ProductRepository
type MockProductRepository struct {
	mu       sync.RWMutex
	products map[string]*domain.Product
}

// NewMockProductRepository creates a new mock product repository
func NewMockProductRepository() *MockProductRepository {
	return &MockProductRepository{
		products: make(map[string]*domain.Product),
	}
}

// Create stores a new product
func (m *MockProductRepository) Create(ctx context.Context, product *domain.Product) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.products[product.ID] = product
	return nil
}

// GetByID retrieves a product by its ID
func (m *MockProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	product, exists := m.products[id]
	if !exists {
		return nil, domain.ErrProductNotFound
	}

	return product, nil
}

// Update updates an existing product
func (m *MockProductRepository) Update(ctx context.Context, product *domain.Product) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.products[product.ID]; !exists {
		return domain.ErrProductNotFound
	}

	m.products[product.ID] = product
	return nil
}

// Delete soft deletes a product
func (m *MockProductRepository) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	product, exists := m.products[id]
	if !exists {
		return domain.ErrProductNotFound
	}

	product.Active = false
	return nil
}

// List retrieves products with pagination and filtering
func (m *MockProductRepository) List(ctx context.Context, filters ports.ListProductFilters) ([]*domain.Product, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var products []*domain.Product
	for _, p := range m.products {
		if filters.ActiveOnly && !p.Active {
			continue
		}
		products = append(products, p)
	}

	return products, int64(len(products)), nil
}

// ExistsByName checks if a product with the given name exists
func (m *MockProductRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.products {
		if p.Name == name {
			return true, nil
		}
	}

	return false, nil
}
