package domain

import (
	"strings"
	"time"
)

// Cart represents a shopping cart (aggregate root)
// This is the main domain entity that contains cart items
type Cart struct {
	UserID    string
	Items     []*CartItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewCart creates a new empty cart for a user
func NewCart(userID string) (*Cart, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}

	now := time.Now()
	return &Cart{
		UserID:    userID,
		Items:     make([]*CartItem, 0),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// AddItem adds a new item to the cart or increments quantity if it already exists
func (c *Cart) AddItem(productID, name string, priceCents, quantity int, imageURL *string) error {
	// Check if item already exists in cart
	for _, item := range c.Items {
		if item.ProductID == productID {
			// Item exists, increment quantity
			if err := item.IncrementQuantity(quantity); err != nil {
				return err
			}
			c.UpdatedAt = time.Now()
			return nil
		}
	}

	// Item doesn't exist, create new cart item
	newItem, err := NewCartItem(productID, name, priceCents, quantity, imageURL)
	if err != nil {
		return err
	}

	c.Items = append(c.Items, newItem)
	c.UpdatedAt = time.Now()
	return nil
}

// RemoveItem removes an item from the cart by product ID
func (c *Cart) RemoveItem(productID string) error {
	if strings.TrimSpace(productID) == "" {
		return ErrInvalidProductID
	}

	for i, item := range c.Items {
		if item.ProductID == productID {
			// Remove item by creating a new slice without this element
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			c.UpdatedAt = time.Now()
			return nil
		}
	}

	return ErrItemNotFound
}

// UpdateItemQuantity updates the quantity of a specific item
func (c *Cart) UpdateItemQuantity(productID string, quantity int) error {
	if strings.TrimSpace(productID) == "" {
		return ErrInvalidProductID
	}

	for _, item := range c.Items {
		if item.ProductID == productID {
			if err := item.UpdateQuantity(quantity); err != nil {
				return err
			}
			c.UpdatedAt = time.Now()
			return nil
		}
	}

	return ErrItemNotFound
}

// Clear removes all items from the cart
func (c *Cart) Clear() {
	c.Items = make([]*CartItem, 0)
	c.UpdatedAt = time.Now()
}

// IsEmpty checks if the cart has no items
func (c *Cart) IsEmpty() bool {
	return len(c.Items) == 0
}

// ItemCount returns the total number of items in the cart
func (c *Cart) ItemCount() int {
	return len(c.Items)
}

// TotalQuantity returns the total quantity of all items
func (c *Cart) TotalQuantity() int {
	total := 0
	for _, item := range c.Items {
		total += item.Quantity
	}
	return total
}

// TotalPrice calculates the total price of all items in the cart
func (c *Cart) TotalPrice() int {
	total := 0
	for _, item := range c.Items {
		total += item.Subtotal()
	}
	return total
}

// GetItem retrieves a specific item by product ID
func (c *Cart) GetItem(productID string) (*CartItem, error) {
	for _, item := range c.Items {
		if item.ProductID == productID {
			return item, nil
		}
	}
	return nil, ErrItemNotFound
}

// HasItem checks if a product is in the cart
func (c *Cart) HasItem(productID string) bool {
	_, err := c.GetItem(productID)
	return err == nil
}
