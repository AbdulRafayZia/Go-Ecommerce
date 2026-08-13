package mocks

import (
	"context"
	"sync"

	"gocommerce/services/product/internal/domain"
)

// MockCategoryRepository is a mock implementation of ports.CategoryRepository
type MockCategoryRepository struct {
	mu         sync.RWMutex
	categories map[string]*domain.Category
}

// NewMockCategoryRepository creates a new mock category repository
func NewMockCategoryRepository() *MockCategoryRepository {
	return &MockCategoryRepository{
		categories: make(map[string]*domain.Category),
	}
}

// Create stores a new category
func (m *MockCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.categories[category.ID] = category
	return nil
}

// GetByID retrieves a category by its ID
func (m *MockCategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	category, exists := m.categories[id]
	if !exists {
		return nil, domain.ErrCategoryNotFound
	}

	return category, nil
}

// Update updates an existing category
func (m *MockCategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.categories[category.ID]; !exists {
		return domain.ErrCategoryNotFound
	}

	m.categories[category.ID] = category
	return nil
}

// Delete deletes a category
func (m *MockCategoryRepository) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.categories[id]; !exists {
		return domain.ErrCategoryNotFound
	}

	delete(m.categories, id)
	return nil
}

// List retrieves all categories
func (m *MockCategoryRepository) List(ctx context.Context) ([]*domain.Category, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var categories []*domain.Category
	for _, c := range m.categories {
		categories = append(categories, c)
	}

	return categories, nil
}

// ExistsByName checks if a category with the given name exists
func (m *MockCategoryRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, c := range m.categories {
		if c.Name == name {
			return true, nil
		}
	}

	return false, nil
}

// GetByParentID retrieves all categories with the given parent ID
func (m *MockCategoryRepository) GetByParentID(ctx context.Context, parentID string) ([]*domain.Category, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var categories []*domain.Category
	for _, c := range m.categories {
		if c.ParentID != nil && *c.ParentID == parentID {
			categories = append(categories, c)
		}
	}

	return categories, nil
}
