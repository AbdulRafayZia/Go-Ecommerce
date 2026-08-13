package domain

import "errors"

// Domain errors - these are business logic errors
var (
	// ErrProductNotFound is returned when a product doesn't exist
	ErrProductNotFound = errors.New("product not found")

	// ErrProductAlreadyExists is returned when creating a product that already exists
	ErrProductAlreadyExists = errors.New("product already exists")

	// ErrInvalidProductData is returned when product data is invalid
	ErrInvalidProductData = errors.New("invalid product data")

	// ErrCategoryNotFound is returned when a category doesn't exist
	ErrCategoryNotFound = errors.New("category not found")

	// ErrCategoryAlreadyExists is returned when creating a category that already exists
	ErrCategoryAlreadyExists = errors.New("category already exists")

	// ErrInvalidPrice is returned when price is invalid
	ErrInvalidPrice = errors.New("price must be non-negative")

	// ErrInvalidStock is returned when stock is invalid
	ErrInvalidStock = errors.New("stock must be non-negative")

	// ErrEmptyProductName is returned when product name is empty
	ErrEmptyProductName = errors.New("product name cannot be empty")

	// ErrEmptyCategoryName is returned when category name is empty
	ErrEmptyCategoryName = errors.New("category name cannot be empty")
)
