package domain

import (
	"strings"
	"time"
)

// Product represents a product in the catalog
// This is a pure domain entity with no infrastructure dependencies
type Product struct {
	ID          string
	Name        string
	Description string
	PriceCents  int
	Currency    string
	CategoryID  *string
	Stock       int
	Active      bool
	ImageURL    *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewProduct creates a new product with validation
func NewProduct(name, description string, priceCents int, currency string, stock int) (*Product, error) {
	// Validate name
	if strings.TrimSpace(name) == "" {
		return nil, ErrEmptyProductName
	}

	// Validate price
	if priceCents < 0 {
		return nil, ErrInvalidPrice
	}

	// Validate stock
	if stock < 0 {
		return nil, ErrInvalidStock
	}

	// Set default currency if empty
	if currency == "" {
		currency = "USD"
	}

	now := time.Now()

	return &Product{
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		PriceCents:  priceCents,
		Currency:    strings.ToUpper(currency),
		Stock:       stock,
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// UpdateName updates the product name with validation
func (p *Product) UpdateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrEmptyProductName
	}
	p.Name = strings.TrimSpace(name)
	p.UpdatedAt = time.Now()
	return nil
}

// UpdateDescription updates the product description
func (p *Product) UpdateDescription(description string) {
	p.Description = strings.TrimSpace(description)
	p.UpdatedAt = time.Now()
}

// UpdatePrice updates the product price with validation
func (p *Product) UpdatePrice(priceCents int) error {
	if priceCents < 0 {
		return ErrInvalidPrice
	}
	p.PriceCents = priceCents
	p.UpdatedAt = time.Now()
	return nil
}

// UpdateStock updates the product stock with validation
func (p *Product) UpdateStock(stock int) error {
	if stock < 0 {
		return ErrInvalidStock
	}
	p.Stock = stock
	p.UpdatedAt = time.Now()
	return nil
}

// AddStock adds to the current stock
func (p *Product) AddStock(quantity int) error {
	if quantity < 0 {
		return ErrInvalidStock
	}
	p.Stock += quantity
	p.UpdatedAt = time.Now()
	return nil
}

// RemoveStock removes from the current stock
func (p *Product) RemoveStock(quantity int) error {
	if quantity < 0 {
		return ErrInvalidStock
	}
	if p.Stock < quantity {
		return ErrInvalidStock
	}
	p.Stock -= quantity
	p.UpdatedAt = time.Now()
	return nil
}

// SetCategory sets the product category
func (p *Product) SetCategory(categoryID *string) {
	p.CategoryID = categoryID
	p.UpdatedAt = time.Now()
}

// SetImageURL sets the product image URL
func (p *Product) SetImageURL(imageURL *string) {
	p.ImageURL = imageURL
	p.UpdatedAt = time.Now()
}

// Deactivate marks the product as inactive (soft delete)
func (p *Product) Deactivate() {
	p.Active = false
	p.UpdatedAt = time.Now()
}

// Activate marks the product as active
func (p *Product) Activate() {
	p.Active = true
	p.UpdatedAt = time.Now()
}

// IsAvailable checks if the product is available for purchase
func (p *Product) IsAvailable() bool {
	return p.Active && p.Stock > 0
}

// CanFulfillQuantity checks if we have enough stock for a quantity
func (p *Product) CanFulfillQuantity(quantity int) bool {
	return p.Active && p.Stock >= quantity
}
