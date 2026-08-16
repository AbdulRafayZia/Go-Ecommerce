package application

import (
	"context"
	"fmt"
	"time"

	"gocommerce/services/cart/internal/domain"
	"gocommerce/services/cart/internal/ports"
)

// Default TTL for carts (7 days)
const DefaultCartTTL = 7 * 24 * time.Hour

// CartService orchestrates cart-related use cases
type CartService struct {
	cartRepo ports.CartRepository
	cartTTL  time.Duration
}

// NewCartService creates a new cart service
func NewCartService(cartRepo ports.CartRepository) *CartService {
	return &CartService{
		cartRepo: cartRepo,
		cartTTL:  DefaultCartTTL,
	}
}

// NewCartServiceWithTTL creates a new cart service with custom TTL
func NewCartServiceWithTTL(cartRepo ports.CartRepository, ttl time.Duration) *CartService {
	return &CartService{
		cartRepo: cartRepo,
		cartTTL:  ttl,
	}
}

// AddItemDTO represents the input for adding an item to cart
type AddItemDTO struct {
	ProductID  string
	Name       string
	PriceCents int
	Quantity   int
	ImageURL   *string
}

// GetCart retrieves a user's cart
func (s *CartService) GetCart(ctx context.Context, userID string) (*domain.Cart, error) {
	cart, err := s.cartRepo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	return cart, nil
}

// AddItem adds an item to the cart or increments quantity if it already exists
// If the cart doesn't exist, it creates a new one
func (s *CartService) AddItem(ctx context.Context, userID string, dto AddItemDTO) (*domain.Cart, error) {
	// Try to get existing cart
	cart, err := s.cartRepo.Get(ctx, userID)
	if err == domain.ErrCartNotFound {
		// Cart doesn't exist, create new one
		cart, err = domain.NewCart(userID)
		if err != nil {
			return nil, fmt.Errorf("failed to create cart: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	// Add item to cart (domain logic handles increment if exists)
	if err := cart.AddItem(dto.ProductID, dto.Name, dto.PriceCents, dto.Quantity, dto.ImageURL); err != nil {
		return nil, err
	}

	// Save cart with TTL
	if err := s.cartRepo.Save(ctx, cart, s.cartTTL); err != nil {
		return nil, fmt.Errorf("failed to save cart: %w", err)
	}

	return cart, nil
}

// UpdateItemQuantity updates the quantity of an item in the cart
func (s *CartService) UpdateItemQuantity(ctx context.Context, userID, productID string, quantity int) (*domain.Cart, error) {
	// Get cart
	cart, err := s.cartRepo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update quantity
	if err := cart.UpdateItemQuantity(productID, quantity); err != nil {
		return nil, err
	}

	// Save cart
	if err := s.cartRepo.Save(ctx, cart, s.cartTTL); err != nil {
		return nil, fmt.Errorf("failed to save cart: %w", err)
	}

	return cart, nil
}

// RemoveItem removes an item from the cart
func (s *CartService) RemoveItem(ctx context.Context, userID, productID string) (*domain.Cart, error) {
	// Get cart
	cart, err := s.cartRepo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Remove item
	if err := cart.RemoveItem(productID); err != nil {
		return nil, err
	}

	// Save cart
	if err := s.cartRepo.Save(ctx, cart, s.cartTTL); err != nil {
		return nil, fmt.Errorf("failed to save cart: %w", err)
	}

	return cart, nil
}

// ClearCart removes all items from the cart
func (s *CartService) ClearCart(ctx context.Context, userID string) error {
	// Check if cart exists
	exists, err := s.cartRepo.Exists(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check cart existence: %w", err)
	}

	if !exists {
		return domain.ErrCartNotFound
	}

	// Delete cart
	if err := s.cartRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete cart: %w", err)
	}

	return nil
}
