package application

import (
	"context"
	"fmt"

	"gocommerce/services/order/internal/domain"
	"gocommerce/services/order/internal/ports"

	"github.com/google/uuid"
)

// OrderService orchestrates order-related use cases
type OrderService struct {
	orderRepo ports.OrderRepository
}

// NewOrderService creates a new order service
func NewOrderService(orderRepo ports.OrderRepository) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
	}
}

// CreateOrderDTO represents the input for creating an order
type CreateOrderDTO struct {
	UserID         string
	Items          []OrderItemDTO
	IdempotencyKey string
}

// OrderItemDTO represents an item in the order creation request
type OrderItemDTO struct {
	ProductID      string
	ProductName    string
	Quantity       int
	UnitPriceCents int
}

// UpdateOrderStatusDTO represents the input for updating order status
type UpdateOrderStatusDTO struct {
	Status         domain.OrderStatus
	PaymentID      *string
	TrackingNumber *string
	Reason         *string
}

// CreateOrder creates a new order with idempotency support
func (s *OrderService) CreateOrder(ctx context.Context, dto CreateOrderDTO) (*domain.Order, error) {
	// Check if order already exists (idempotency)
	existingOrder, err := s.orderRepo.GetByIdempotencyKey(ctx, dto.IdempotencyKey)
	if err == nil {
		// Order already exists, return it
		return existingOrder, nil
	}
	if err != domain.ErrOrderNotFound {
		return nil, fmt.Errorf("failed to check order existence: %w", err)
	}

	// Convert DTOs to domain order items
	orderItems := make([]*domain.OrderItem, 0, len(dto.Items))
	for _, itemDTO := range dto.Items {
		item, err := domain.NewOrderItem(
			itemDTO.ProductID,
			itemDTO.ProductName,
			itemDTO.Quantity,
			itemDTO.UnitPriceCents,
		)
		if err != nil {
			return nil, err
		}
		orderItems = append(orderItems, item)
	}

	// Create domain order
	order, err := domain.NewOrder(dto.UserID, dto.IdempotencyKey, orderItems)
	if err != nil {
		return nil, err
	}

	// Generate ID
	order.ID = uuid.New().String()

	// Mark as awaiting payment (first state transition)
	if err := order.MarkAwaitingPayment(); err != nil {
		return nil, fmt.Errorf("failed to mark order as awaiting payment: %w", err)
	}

	// Persist order (this also stores domain events in outbox)
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return order, nil
}

// GetOrder retrieves an order by ID
func (s *OrderService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return order, nil
}

// ListOrders retrieves orders with pagination and filtering
func (s *OrderService) ListOrders(ctx context.Context, userID *string, status *domain.OrderStatus, page, pageSize int) ([]*domain.Order, int64, error) {
	filters := ports.ListOrderFilters{
		Page:     page,
		PageSize: pageSize,
		UserID:   userID,
		Status:   status,
	}

	orders, total, err := s.orderRepo.List(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}

	return orders, total, nil
}

// UpdateOrderStatus updates the order status with appropriate state transitions
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, dto UpdateOrderStatusDTO) (*domain.Order, error) {
	// Get order
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Perform state transition based on target status
	switch dto.Status {
	case domain.OrderStatusPaid:
		if dto.PaymentID == nil {
			return nil, fmt.Errorf("payment_id is required when marking order as paid")
		}
		if err := order.MarkPaid(*dto.PaymentID); err != nil {
			return nil, err
		}

	case domain.OrderStatusFulfilling:
		if err := order.MarkFulfilling(); err != nil {
			return nil, err
		}

	case domain.OrderStatusShipped:
		if err := order.MarkShipped(dto.TrackingNumber); err != nil {
			return nil, err
		}

	case domain.OrderStatusDelivered:
		if err := order.MarkDelivered(); err != nil {
			return nil, err
		}

	case domain.OrderStatusCancelled:
		reason := ""
		if dto.Reason != nil {
			reason = *dto.Reason
		}
		if err := order.Cancel(reason); err != nil {
			return nil, err
		}

	case domain.OrderStatusFailed:
		reason := ""
		if dto.Reason != nil {
			reason = *dto.Reason
		}
		if err := order.MarkFailed(reason); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported status transition to %s", dto.Status)
	}

	// Save updated order (this also stores new domain events in outbox)
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	return order, nil
}

// CancelOrder cancels an order
func (s *OrderService) CancelOrder(ctx context.Context, orderID string, reason string) (*domain.Order, error) {
	return s.UpdateOrderStatus(ctx, orderID, UpdateOrderStatusDTO{
		Status: domain.OrderStatusCancelled,
		Reason: &reason,
	})
}
