package domain

import "strings"

// CartItem represents an item in the shopping cart
// This is a value object - it doesn't have its own identity outside the cart
type CartItem struct {
	ProductID  string
	Name       string
	PriceCents int
	Quantity   int
	ImageURL   *string
}

// NewCartItem creates a new cart item with validation
func NewCartItem(productID, name string, priceCents, quantity int, imageURL *string) (*CartItem, error) {
	// Validate product ID
	if strings.TrimSpace(productID) == "" {
		return nil, ErrInvalidProductID
	}

	// Validate quantity
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	// Validate price
	if priceCents < 0 {
		return nil, ErrInvalidPrice
	}

	return &CartItem{
		ProductID:  productID,
		Name:       name,
		PriceCents: priceCents,
		Quantity:   quantity,
		ImageURL:   imageURL,
	}, nil
}

// Subtotal calculates the subtotal for this cart item
func (i *CartItem) Subtotal() int {
	return i.PriceCents * i.Quantity
}

// UpdateQuantity updates the quantity of this cart item
func (i *CartItem) UpdateQuantity(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	i.Quantity = quantity
	return nil
}

// IncrementQuantity increases the quantity by the specified amount
func (i *CartItem) IncrementQuantity(amount int) error {
	if amount <= 0 {
		return ErrInvalidQuantity
	}
	i.Quantity += amount
	return nil
}
