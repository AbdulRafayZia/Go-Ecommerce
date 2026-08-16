package domain

import (
	"testing"
)

func TestNewCart(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		wantErr error
	}{
		{
			name:    "valid cart creation",
			userID:  "user-123",
			wantErr: nil,
		},
		{
			name:    "empty user ID",
			userID:  "",
			wantErr: ErrInvalidUserID,
		},
		{
			name:    "whitespace user ID",
			userID:  "   ",
			wantErr: ErrInvalidUserID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cart, err := NewCart(tt.userID)

			if err != tt.wantErr {
				t.Errorf("NewCart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if cart.UserID != tt.userID {
					t.Errorf("NewCart() userID = %v, want %v", cart.UserID, tt.userID)
				}
				if cart.IsEmpty() != true {
					t.Errorf("NewCart() should create empty cart")
				}
				if cart.CreatedAt.IsZero() {
					t.Errorf("NewCart() should set CreatedAt")
				}
			}
		})
	}
}

func TestCart_AddItem(t *testing.T) {
	tests := []struct {
		name       string
		productID  string
		itemName   string
		priceCents int
		quantity   int
		wantErr    error
	}{
		{
			name:       "add valid item",
			productID:  "prod-1",
			itemName:   "Product 1",
			priceCents: 1000,
			quantity:   2,
			wantErr:    nil,
		},
		{
			name:       "add item with zero price",
			productID:  "prod-2",
			itemName:   "Free Product",
			priceCents: 0,
			quantity:   1,
			wantErr:    nil,
		},
		{
			name:       "add item with invalid quantity",
			productID:  "prod-3",
			itemName:   "Product 3",
			priceCents: 500,
			quantity:   0,
			wantErr:    ErrInvalidQuantity,
		},
		{
			name:       "add item with negative price",
			productID:  "prod-4",
			itemName:   "Product 4",
			priceCents: -100,
			quantity:   1,
			wantErr:    ErrInvalidPrice,
		},
		{
			name:       "add item with empty product ID",
			productID:  "",
			itemName:   "Product 5",
			priceCents: 100,
			quantity:   1,
			wantErr:    ErrInvalidProductID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cart, _ := NewCart("user-123")
			err := cart.AddItem(tt.productID, tt.itemName, tt.priceCents, tt.quantity, nil)

			if err != tt.wantErr {
				t.Errorf("AddItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if cart.ItemCount() != 1 {
					t.Errorf("AddItem() item count = %v, want 1", cart.ItemCount())
				}
				item, _ := cart.GetItem(tt.productID)
				if item.Quantity != tt.quantity {
					t.Errorf("AddItem() quantity = %v, want %v", item.Quantity, tt.quantity)
				}
			}
		})
	}
}

func TestCart_AddItem_IncrementExisting(t *testing.T) {
	cart, _ := NewCart("user-123")

	// Add item first time
	err := cart.AddItem("prod-1", "Product 1", 1000, 2, nil)
	if err != nil {
		t.Fatalf("First AddItem() failed: %v", err)
	}

	// Add same item again
	err = cart.AddItem("prod-1", "Product 1", 1000, 3, nil)
	if err != nil {
		t.Fatalf("Second AddItem() failed: %v", err)
	}

	// Should have only 1 item with quantity 5
	if cart.ItemCount() != 1 {
		t.Errorf("ItemCount() = %v, want 1", cart.ItemCount())
	}

	item, _ := cart.GetItem("prod-1")
	if item.Quantity != 5 {
		t.Errorf("Quantity after increment = %v, want 5", item.Quantity)
	}
}

func TestCart_RemoveItem(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*Cart)
		productID string
		wantErr   error
	}{
		{
			name: "remove existing item",
			setupFunc: func(c *Cart) {
				c.AddItem("prod-1", "Product 1", 1000, 2, nil)
			},
			productID: "prod-1",
			wantErr:   nil,
		},
		{
			name: "remove non-existent item",
			setupFunc: func(c *Cart) {
				c.AddItem("prod-1", "Product 1", 1000, 2, nil)
			},
			productID: "prod-999",
			wantErr:   ErrItemNotFound,
		},
		{
			name:      "remove from empty cart",
			setupFunc: func(c *Cart) {},
			productID: "prod-1",
			wantErr:   ErrItemNotFound,
		},
		{
			name: "remove with empty product ID",
			setupFunc: func(c *Cart) {
				c.AddItem("prod-1", "Product 1", 1000, 2, nil)
			},
			productID: "",
			wantErr:   ErrInvalidProductID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cart, _ := NewCart("user-123")
			tt.setupFunc(cart)

			initialCount := cart.ItemCount()
			err := cart.RemoveItem(tt.productID)

			if err != tt.wantErr {
				t.Errorf("RemoveItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && cart.ItemCount() != initialCount-1 {
				t.Errorf("RemoveItem() item count = %v, want %v", cart.ItemCount(), initialCount-1)
			}
		})
	}
}

func TestCart_UpdateItemQuantity(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*Cart)
		productID   string
		newQuantity int
		wantErr     error
	}{
		{
			name: "update existing item quantity",
			setupFunc: func(c *Cart) {
				c.AddItem("prod-1", "Product 1", 1000, 2, nil)
			},
			productID:   "prod-1",
			newQuantity: 5,
			wantErr:     nil,
		},
		{
			name: "update with zero quantity",
			setupFunc: func(c *Cart) {
				c.AddItem("prod-1", "Product 1", 1000, 2, nil)
			},
			productID:   "prod-1",
			newQuantity: 0,
			wantErr:     ErrInvalidQuantity,
		},
		{
			name: "update non-existent item",
			setupFunc: func(c *Cart) {
				c.AddItem("prod-1", "Product 1", 1000, 2, nil)
			},
			productID:   "prod-999",
			newQuantity: 5,
			wantErr:     ErrItemNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cart, _ := NewCart("user-123")
			tt.setupFunc(cart)

			err := cart.UpdateItemQuantity(tt.productID, tt.newQuantity)

			if err != tt.wantErr {
				t.Errorf("UpdateItemQuantity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				item, _ := cart.GetItem(tt.productID)
				if item.Quantity != tt.newQuantity {
					t.Errorf("UpdateItemQuantity() quantity = %v, want %v", item.Quantity, tt.newQuantity)
				}
			}
		})
	}
}

func TestCart_Clear(t *testing.T) {
	cart, _ := NewCart("user-123")
	cart.AddItem("prod-1", "Product 1", 1000, 2, nil)
	cart.AddItem("prod-2", "Product 2", 2000, 1, nil)

	if cart.ItemCount() != 2 {
		t.Fatalf("Setup failed: cart should have 2 items")
	}

	cart.Clear()

	if !cart.IsEmpty() {
		t.Errorf("Clear() cart should be empty")
	}

	if cart.ItemCount() != 0 {
		t.Errorf("Clear() item count = %v, want 0", cart.ItemCount())
	}
}

func TestCart_Calculations(t *testing.T) {
	cart, _ := NewCart("user-123")
	cart.AddItem("prod-1", "Product 1", 1000, 2, nil)  // 2000
	cart.AddItem("prod-2", "Product 2", 1500, 3, nil)  // 4500
	cart.AddItem("prod-3", "Product 3", 500, 1, nil)   // 500

	tests := []struct {
		name     string
		testFunc func() int
		want     int
	}{
		{
			name:     "item count",
			testFunc: cart.ItemCount,
			want:     3,
		},
		{
			name:     "total quantity",
			testFunc: cart.TotalQuantity,
			want:     6,
		},
		{
			name:     "total price",
			testFunc: cart.TotalPrice,
			want:     7000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.testFunc()
			if got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestCart_HasItem(t *testing.T) {
	cart, _ := NewCart("user-123")
	cart.AddItem("prod-1", "Product 1", 1000, 2, nil)

	if !cart.HasItem("prod-1") {
		t.Errorf("HasItem() = false, want true for existing item")
	}

	if cart.HasItem("prod-999") {
		t.Errorf("HasItem() = true, want false for non-existent item")
	}
}
