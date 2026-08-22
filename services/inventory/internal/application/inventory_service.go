package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"gocommerce/services/inventory/internal/domain"
	"gocommerce/services/inventory/internal/ports"
)

// InventoryService handles inventory operations
type InventoryService struct {
	stockRepo       ports.StockRepository
	reservationRepo ports.ReservationRepository
	db              *sql.DB
}

// NewInventoryService creates a new inventory service
func NewInventoryService(
	stockRepo ports.StockRepository,
	reservationRepo ports.ReservationRepository,
	db *sql.DB,
) *InventoryService {
	return &InventoryService{
		stockRepo:       stockRepo,
		reservationRepo: reservationRepo,
		db:              db,
	}
}

// ReserveStockForOrder creates a reservation and reserves stock for an order
// This operation is transactional to ensure consistency
func (s *InventoryService) ReserveStockForOrder(ctx context.Context, orderID, productID string, quantity int) (*domain.Reservation, error) {
	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if reservation already exists (idempotency)
	existingReservation, err := s.reservationRepo.GetByOrderAndProduct(ctx, orderID, productID)
	if err == nil {
		// Reservation already exists, return it
		return existingReservation, nil
	}
	if err != domain.ErrReservationNotFound {
		return nil, fmt.Errorf("failed to check existing reservation: %w", err)
	}

	// Reserve stock
	if err := s.stockRepo.ReserveStock(ctx, productID, quantity); err != nil {
		return nil, fmt.Errorf("failed to reserve stock: %w", err)
	}

	// Create reservation record
	reservation, err := domain.NewReservation(orderID, productID, quantity)
	if err != nil {
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	reservation.ID = uuid.New().String()

	if err := s.reservationRepo.Create(ctx, reservation); err != nil {
		return nil, fmt.Errorf("failed to save reservation: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return reservation, nil
}

// FulfillReservation marks a reservation as fulfilled and deducts reserved stock
func (s *InventoryService) FulfillReservation(ctx context.Context, orderID string) error {
	// Get all reservations for the order
	reservations, err := s.reservationRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get reservations: %w", err)
	}

	if len(reservations) == 0 {
		return domain.ErrReservationNotFound
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Process each reservation
	for _, reservation := range reservations {
		// Skip if already fulfilled
		if reservation.Status == domain.ReservationStatusFulfilled {
			continue
		}

		// Fulfill the reservation in domain
		if err := reservation.Fulfill(); err != nil {
			return fmt.Errorf("failed to fulfill reservation: %w", err)
		}

		// Update reservation status
		if err := s.reservationRepo.Update(ctx, reservation); err != nil {
			return fmt.Errorf("failed to update reservation: %w", err)
		}

		// Deduct from reserved stock
		if err := s.stockRepo.FulfillReservation(ctx, reservation.ProductID, reservation.Quantity); err != nil {
			return fmt.Errorf("failed to fulfill stock reservation: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CancelReservation marks a reservation as cancelled and releases reserved stock
func (s *InventoryService) CancelReservation(ctx context.Context, orderID string) error {
	// Get all reservations for the order
	reservations, err := s.reservationRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get reservations: %w", err)
	}

	if len(reservations) == 0 {
		// No reservations found, nothing to cancel (could be already cancelled or never created)
		return nil
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Process each reservation
	for _, reservation := range reservations {
		// Skip if already in final state
		if reservation.IsFinal() {
			continue
		}

		// Cancel the reservation in domain
		if err := reservation.Cancel(); err != nil {
			return fmt.Errorf("failed to cancel reservation: %w", err)
		}

		// Update reservation status
		if err := s.reservationRepo.Update(ctx, reservation); err != nil {
			return fmt.Errorf("failed to update reservation: %w", err)
		}

		// Release reserved stock back to available
		if err := s.stockRepo.ReleaseReservation(ctx, reservation.ProductID, reservation.Quantity); err != nil {
			return fmt.Errorf("failed to release stock reservation: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ExpireOldReservations expires reservations older than the given duration
// This should be run periodically to clean up abandoned reservations
func (s *InventoryService) ExpireOldReservations(ctx context.Context, olderThan time.Duration) error {
	reservations, err := s.reservationRepo.ListExpiredReservations(ctx, olderThan)
	if err != nil {
		return fmt.Errorf("failed to list expired reservations: %w", err)
	}

	for _, reservation := range reservations {
		// Start transaction for each reservation
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Expire the reservation
		if err := reservation.Expire(); err != nil {
			tx.Rollback()
			continue // Skip this one, continue with others
		}

		// Update reservation status
		if err := s.reservationRepo.Update(ctx, reservation); err != nil {
			tx.Rollback()
			continue
		}

		// Release reserved stock
		if err := s.stockRepo.ReleaseReservation(ctx, reservation.ProductID, reservation.Quantity); err != nil {
			tx.Rollback()
			continue
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			continue
		}
	}

	return nil
}

// GetStock retrieves stock information for a product
func (s *InventoryService) GetStock(ctx context.Context, productID string) (*domain.Stock, error) {
	return s.stockRepo.GetByProductID(ctx, productID)
}

// AddStock adds new stock for a product (restocking)
func (s *InventoryService) AddStock(ctx context.Context, productID string, quantity int) error {
	stock, err := s.stockRepo.GetByProductID(ctx, productID)
	if err == domain.ErrStockNotFound {
		// Create new stock record
		stock, err = domain.NewStock(productID, quantity)
		if err != nil {
			return fmt.Errorf("failed to create stock: %w", err)
		}
		return s.stockRepo.Create(ctx, stock)
	}
	if err != nil {
		return fmt.Errorf("failed to get stock: %w", err)
	}

	// Add to existing stock
	if err := stock.AddStock(quantity); err != nil {
		return fmt.Errorf("failed to add stock: %w", err)
	}

	return s.stockRepo.Update(ctx, stock)
}

// SetStock sets the available stock for a product (inventory adjustment)
func (s *InventoryService) SetStock(ctx context.Context, productID string, quantity int) error {
	stock, err := s.stockRepo.GetByProductID(ctx, productID)
	if err == domain.ErrStockNotFound {
		// Create new stock record
		stock, err = domain.NewStock(productID, quantity)
		if err != nil {
			return fmt.Errorf("failed to create stock: %w", err)
		}
		return s.stockRepo.Create(ctx, stock)
	}
	if err != nil {
		return fmt.Errorf("failed to get stock: %w", err)
	}

	// Set stock quantity
	if err := stock.SetStock(quantity); err != nil {
		return fmt.Errorf("failed to set stock: %w", err)
	}

	return s.stockRepo.Update(ctx, stock)
}

// ListLowStock returns products with stock below the threshold
func (s *InventoryService) ListLowStock(ctx context.Context, threshold int) ([]*domain.Stock, error) {
	return s.stockRepo.ListLowStock(ctx, threshold)
}
