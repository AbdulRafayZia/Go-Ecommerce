package domain

import "testing"

func TestNewCartItem(t *testing.T) {
	tests := []struct {
		name       string
		productID  string
		itemName   string
		priceCents int
		quantity   int
		wantErr    error
	}{
		{
			name:       "valid cart item",
			productID:  "prod-1",
			itemName:   "Product 1",
			priceCents: 1000,
			quantity:   2,
			wantErr:    nil,
		},
		{
			name:       "cart item with zero price",
			productID:  "prod-2",
			itemName:   "Free Product",
			priceCents: 0,
			quantity:   1,
			wantErr:    nil,
		},
		{
			name:       "empty product ID",
			productID:  "",
			itemName:   "Product",
			priceCents: 100,
			quantity:   1,
			wantErr:    ErrInvalidProductID,
		},
		{
			name:       "whitespace product ID",
			productID:  "   ",
			itemName:   "Product",
			priceCents: 100,
			quantity:   1,
			wantErr:    ErrInvalidProductID,
		},
		{
			name:       "zero quantity",
			productID:  "prod-1",
			itemName:   "Product",
			priceCents: 100,
			quantity:   0,
			wantErr:    ErrInvalidQuantity,
		},
		{
			name:       "negative quantity",
			productID:  "prod-1",
			itemName:   "Product",
			priceCents: 100,
			quantity:   -1,
			wantErr:    ErrInvalidQuantity,
		},
		{
			name:       "negative price",
			productID:  "prod-1",
			itemName:   "Product",
			priceCents: -100,
			quantity:   1,
			wantErr:    ErrInvalidPrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := NewCartItem(tt.productID, tt.itemName, tt.priceCents, tt.quantity, nil)

			if err != tt.wantErr {
				t.Errorf("NewCartItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if item.ProductID != tt.productID {
					t.Errorf("NewCartItem() productID = %v, want %v", item.ProductID, tt.productID)
				}
				if item.Name != tt.itemName {
					t.Errorf("NewCartItem() name = %v, want %v", item.Name, tt.itemName)
				}
				if item.PriceCents != tt.priceCents {
					t.Errorf("NewCartItem() priceCents = %v, want %v", item.PriceCents, tt.priceCents)
				}
				if item.Quantity != tt.quantity {
					t.Errorf("NewCartItem() quantity = %v, want %v", item.Quantity, tt.quantity)
				}
			}
		})
	}
}

func TestCartItem_Subtotal(t *testing.T) {
	tests := []struct {
		name       string
		priceCents int
		quantity   int
		want       int
	}{
		{
			name:       "normal calculation",
			priceCents: 1000,
			quantity:   3,
			want:       3000,
		},
		{
			name:       "single item",
			priceCents: 500,
			quantity:   1,
			want:       500,
		},
		{
			name:       "zero price",
			priceCents: 0,
			quantity:   5,
			want:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, _ := NewCartItem("prod-1", "Product", tt.priceCents, tt.quantity, nil)
			got := item.Subtotal()

			if got != tt.want {
				t.Errorf("Subtotal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCartItem_UpdateQuantity(t *testing.T) {
	tests := []struct {
		name        string
		newQuantity int
		wantErr     error
	}{
		{
			name:        "update to valid quantity",
			newQuantity: 5,
			wantErr:     nil,
		},
		{
			name:        "update to zero",
			newQuantity: 0,
			wantErr:     ErrInvalidQuantity,
		},
		{
			name:        "update to negative",
			newQuantity: -1,
			wantErr:     ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, _ := NewCartItem("prod-1", "Product", 1000, 2, nil)
			err := item.UpdateQuantity(tt.newQuantity)

			if err != tt.wantErr {
				t.Errorf("UpdateQuantity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && item.Quantity != tt.newQuantity {
				t.Errorf("UpdateQuantity() quantity = %v, want %v", item.Quantity, tt.newQuantity)
			}
		})
	}
}

func TestCartItem_IncrementQuantity(t *testing.T) {
	tests := []struct {
		name           string
		initialQty     int
		incrementBy    int
		wantQuantity   int
		wantErr        error
	}{
		{
			name:         "increment by positive amount",
			initialQty:   2,
			incrementBy:  3,
			wantQuantity: 5,
			wantErr:      nil,
		},
		{
			name:         "increment by zero",
			initialQty:   2,
			incrementBy:  0,
			wantQuantity: 2,
			wantErr:      ErrInvalidQuantity,
		},
		{
			name:         "increment by negative",
			initialQty:   5,
			incrementBy:  -2,
			wantQuantity: 5,
			wantErr:      ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, _ := NewCartItem("prod-1", "Product", 1000, tt.initialQty, nil)
			err := item.IncrementQuantity(tt.incrementBy)

			if err != tt.wantErr {
				t.Errorf("IncrementQuantity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if item.Quantity != tt.wantQuantity {
				t.Errorf("IncrementQuantity() quantity = %v, want %v", item.Quantity, tt.wantQuantity)
			}
		})
	}
}
