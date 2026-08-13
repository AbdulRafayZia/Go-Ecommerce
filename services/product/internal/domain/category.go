package domain

import (
	"strings"
	"time"
)

// Category represents a product category
type Category struct {
	ID          string
	Name        string
	Description *string
	ParentID    *string
	CreatedAt   time.Time
}

// NewCategory creates a new category with validation
func NewCategory(name string, description *string, parentID *string) (*Category, error) {
	// Validate name
	if strings.TrimSpace(name) == "" {
		return nil, ErrEmptyCategoryName
	}

	return &Category{
		Name:        strings.TrimSpace(name),
		Description: description,
		ParentID:    parentID,
		CreatedAt:   time.Now(),
	}, nil
}

// UpdateName updates the category name with validation
func (c *Category) UpdateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrEmptyCategoryName
	}
	c.Name = strings.TrimSpace(name)
	return nil
}

// UpdateDescription updates the category description
func (c *Category) UpdateDescription(description *string) {
	c.Description = description
}

// HasParent checks if the category has a parent category
func (c *Category) HasParent() bool {
	return c.ParentID != nil && *c.ParentID != ""
}
