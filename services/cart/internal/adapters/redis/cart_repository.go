package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gocommerce/services/cart/internal/domain"
	"gocommerce/services/cart/internal/ports"

	"github.com/redis/go-redis/v9"
)

// CartRepository is the Redis implementation of ports.CartRepository
type CartRepository struct {
	client *redis.Client
}

// NewCartRepository creates a new Redis cart repository
func NewCartRepository(client *redis.Client) ports.CartRepository {
	return &CartRepository{
		client: client,
	}
}

// cartKey generates the Redis key for a user's cart
func (r *CartRepository) cartKey(userID string) string {
	return fmt.Sprintf("cart:%s", userID)
}

// Save stores a cart in Redis with TTL
func (r *CartRepository) Save(ctx context.Context, cart *domain.Cart, ttl time.Duration) error {
	key := r.cartKey(cart.UserID)

	// Serialize cart to JSON
	data, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("failed to marshal cart: %w", err)
	}

	// Store in Redis with TTL
	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to save cart: %w", err)
	}

	return nil
}

// Get retrieves a cart from Redis
func (r *CartRepository) Get(ctx context.Context, userID string) (*domain.Cart, error) {
	key := r.cartKey(userID)

	// Get from Redis
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, domain.ErrCartNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	// Deserialize cart from JSON
	var cart domain.Cart
	if err := json.Unmarshal(data, &cart); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cart: %w", err)
	}

	return &cart, nil
}

// Delete removes a cart from Redis
func (r *CartRepository) Delete(ctx context.Context, userID string) error {
	key := r.cartKey(userID)

	result, err := r.client.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to delete cart: %w", err)
	}

	if result == 0 {
		return domain.ErrCartNotFound
	}

	return nil
}

// Exists checks if a cart exists in Redis
func (r *CartRepository) Exists(ctx context.Context, userID string) (bool, error) {
	key := r.cartKey(userID)

	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cart existence: %w", err)
	}

	return result > 0, nil
}
