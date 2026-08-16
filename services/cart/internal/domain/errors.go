package domain

import "errors"

// Domain errors for the cart service
var (
	// ErrCartNotFound is returned when a cart is not found
	ErrCartNotFound = errors.New("cart not found")

	// ErrInvalidQuantity is returned when an invalid quantity is provided
	ErrInvalidQuantity = errors.New("quantity must be greater than zero")

	// ErrItemNotFound is returned when an item is not found in the cart
	ErrItemNotFound = errors.New("item not found in cart")

	// ErrEmptyCart is returned when trying to perform operations on an empty cart
	ErrEmptyCart = errors.New("cart is empty")

	// ErrInvalidUserID is returned when an invalid user ID is provided
	ErrInvalidUserID = errors.New("user ID cannot be empty")

	// ErrInvalidProductID is returned when an invalid product ID is provided
	ErrInvalidProductID = errors.New("product ID cannot be empty")

	// ErrInvalidPrice is returned when an invalid price is provided
	ErrInvalidPrice = errors.New("price must be greater than or equal to zero")
)
