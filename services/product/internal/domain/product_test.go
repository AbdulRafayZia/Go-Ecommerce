package domain

import (
	"testing"
)

func TestNewProduct(t *testing.T) {
	tests := []struct {
		name        string
		productName string
		description string
		priceCents  int
		currency    string
		stock       int
		wantErr     error
	}{
		{
			name:        "valid product",
			productName: "Test Product",
			description: "A test product",
			priceCents:  1000,
			currency:    "USD",
			stock:       10,
			wantErr:     nil,
		},
		{
			name:        "empty name",
			productName: "",
			description: "A test product",
			priceCents:  1000,
			currency:    "USD",
			stock:       10,
			wantErr:     ErrEmptyProductName,
		},
		{
			name:        "whitespace name",
			productName: "   ",
			description: "A test product",
			priceCents:  1000,
			currency:    "USD",
			stock:       10,
			wantErr:     ErrEmptyProductName,
		},
		{
			name:        "negative price",
			productName: "Test Product",
			description: "A test product",
			priceCents:  -100,
			currency:    "USD",
			stock:       10,
			wantErr:     ErrInvalidPrice,
		},
		{
			name:        "zero price is valid",
			productName: "Free Product",
			description: "A free product",
			priceCents:  0,
			currency:    "USD",
			stock:       10,
			wantErr:     nil,
		},
		{
			name:        "negative stock",
			productName: "Test Product",
			description: "A test product",
			priceCents:  1000,
			currency:    "USD",
			stock:       -5,
			wantErr:     ErrInvalidStock,
		},
		{
			name:        "zero stock is valid",
			productName: "Out of Stock Product",
			description: "A product with no stock",
			priceCents:  1000,
			currency:    "USD",
			stock:       0,
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := NewProduct(tt.productName, tt.description, tt.priceCents, tt.currency, tt.stock)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("NewProduct() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("NewProduct() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("NewProduct() unexpected error = %v", err)
				return
			}

			if product == nil {
				t.Error("NewProduct() returned nil product")
				return
			}

			if product.Name != tt.productName {
				t.Errorf("Product.Name = %v, want %v", product.Name, tt.productName)
			}

			if product.Description != tt.description {
				t.Errorf("Product.Description = %v, want %v", product.Description, tt.description)
			}

			if product.PriceCents != tt.priceCents {
				t.Errorf("Product.PriceCents = %v, want %v", product.PriceCents, tt.priceCents)
			}

			if product.Currency != tt.currency {
				t.Errorf("Product.Currency = %v, want %v", product.Currency, tt.currency)
			}

			if product.Stock != tt.stock {
				t.Errorf("Product.Stock = %v, want %v", product.Stock, tt.stock)
			}

			if !product.Active {
				t.Error("Product.Active should be true for new products")
			}
		})
	}
}

func TestProduct_UpdatePrice(t *testing.T) {
	product, _ := NewProduct("Test Product", "Description", 1000, "USD", 10)

	tests := []struct {
		name       string
		priceCents int
		wantErr    error
	}{
		{
			name:       "valid price update",
			priceCents: 2000,
			wantErr:    nil,
		},
		{
			name:       "zero price is valid",
			priceCents: 0,
			wantErr:    nil,
		},
		{
			name:       "negative price",
			priceCents: -100,
			wantErr:    ErrInvalidPrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := product.UpdatePrice(tt.priceCents)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("UpdatePrice() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("UpdatePrice() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("UpdatePrice() unexpected error = %v", err)
				return
			}

			if product.PriceCents != tt.priceCents {
				t.Errorf("Product.PriceCents = %v, want %v", product.PriceCents, tt.priceCents)
			}
		})
	}
}

func TestProduct_UpdateStock(t *testing.T) {
	product, _ := NewProduct("Test Product", "Description", 1000, "USD", 10)

	tests := []struct {
		name    string
		stock   int
		wantErr error
	}{
		{
			name:    "increase stock",
			stock:   20,
			wantErr: nil,
		},
		{
			name:    "decrease stock",
			stock:   5,
			wantErr: nil,
		},
		{
			name:    "zero stock is valid",
			stock:   0,
			wantErr: nil,
		},
		{
			name:    "negative stock",
			stock:   -5,
			wantErr: ErrInvalidStock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := product.UpdateStock(tt.stock)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("UpdateStock() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("UpdateStock() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateStock() unexpected error = %v", err)
				return
			}

			if product.Stock != tt.stock {
				t.Errorf("Product.Stock = %v, want %v", product.Stock, tt.stock)
			}
		})
	}
}

func TestProduct_UpdateName(t *testing.T) {
	product, _ := NewProduct("Test Product", "Description", 1000, "USD", 10)

	tests := []struct {
		name    string
		newName string
		wantErr error
	}{
		{
			name:    "valid name update",
			newName: "Updated Product",
			wantErr: nil,
		},
		{
			name:    "empty name",
			newName: "",
			wantErr: ErrEmptyProductName,
		},
		{
			name:    "whitespace name",
			newName: "   ",
			wantErr: ErrEmptyProductName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := product.UpdateName(tt.newName)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("UpdateName() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("UpdateName() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateName() unexpected error = %v", err)
				return
			}

			if product.Name != tt.newName {
				t.Errorf("Product.Name = %v, want %v", product.Name, tt.newName)
			}
		})
	}
}

func TestProduct_ActivateDeactivate(t *testing.T) {
	product, _ := NewProduct("Test Product", "Description", 1000, "USD", 10)

	if !product.Active {
		t.Error("New product should be active")
	}

	product.Deactivate()
	if product.Active {
		t.Error("Product should be inactive after Deactivate()")
	}

	product.Activate()
	if !product.Active {
		t.Error("Product should be active after Activate()")
	}
}

func TestProduct_SetCategory(t *testing.T) {
	product, _ := NewProduct("Test Product", "Description", 1000, "USD", 10)

	categoryID := "category-123"
	product.SetCategory(&categoryID)

	if product.CategoryID == nil {
		t.Error("CategoryID should not be nil after SetCategory()")
		return
	}

	if *product.CategoryID != categoryID {
		t.Errorf("Product.CategoryID = %v, want %v", *product.CategoryID, categoryID)
	}

	// Test unsetting category
	product.SetCategory(nil)
	if product.CategoryID != nil {
		t.Error("CategoryID should be nil after SetCategory(nil)")
	}
}

func TestProduct_SetImageURL(t *testing.T) {
	product, _ := NewProduct("Test Product", "Description", 1000, "USD", 10)

	imageURL := "https://example.com/image.jpg"
	product.SetImageURL(&imageURL)

	if product.ImageURL == nil {
		t.Error("ImageURL should not be nil after SetImageURL()")
		return
	}

	if *product.ImageURL != imageURL {
		t.Errorf("Product.ImageURL = %v, want %v", *product.ImageURL, imageURL)
	}

	// Test unsetting image URL
	product.SetImageURL(nil)
	if product.ImageURL != nil {
		t.Error("ImageURL should be nil after SetImageURL(nil)")
	}
}

func TestProduct_UpdateDescription(t *testing.T) {
	product, _ := NewProduct("Test Product", "Original Description", 1000, "USD", 10)

	newDescription := "Updated Description"
	product.UpdateDescription(newDescription)

	if product.Description != newDescription {
		t.Errorf("Product.Description = %v, want %v", product.Description, newDescription)
	}

	// Test empty description
	product.UpdateDescription("")
	if product.Description != "" {
		t.Errorf("Product.Description = %v, want empty string", product.Description)
	}
}
