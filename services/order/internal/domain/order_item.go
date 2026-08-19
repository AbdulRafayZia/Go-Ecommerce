package domain

import "strings"

// OrderItem represents an item in an order
// This is a value object - it doesn't have its own identity outside the order
type OrderItem struct {
	ProductID     string
	ProductName   string
	Quantity      int
	UnitPriceCents int
}

// NewOrderItem creates a new order item with validation
func NewOrderItem(productID, productName string, quantity, unitPriceCents int) (*OrderItem, error) {
	// Validate product ID
	if strings.TrimSpace(productID) == "" {
		return nil, ErrInvalidProductID
	}

	// Validate quantity
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	// Validate price
	if unitPriceCents < 0 {
		return nil, ErrInvalidPrice
	}

	return &OrderItem{
		ProductID:      productID,
		ProductName:    productName,
		Quantity:       quantity,
		UnitPriceCents: unitPriceCents,
	}, nil
}

// Subtotal calculates the subtotal for this order item
func (i *OrderItem) Subtotal() int {
	return i.UnitPriceCents * i.Quantity
}
