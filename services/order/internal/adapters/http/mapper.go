package api

import (
	"gocommerce/services/order/internal/application"
	"gocommerce/services/order/internal/domain"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

// domainOrderToAPI converts a domain Order to an API Order response
func domainOrderToAPI(order *domain.Order) Order {
	items := make([]OrderItem, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, domainOrderItemToAPI(item))
	}

	return Order{
		Id:             stringToUUID(order.ID),
		UserId:         stringToUUID(order.UserID),
		Status:         OrderStatus(order.Status),
		Items:          items,
		TotalCents:     order.TotalCents,
		IdempotencyKey: order.IdempotencyKey,
		PaymentId:      stringPtrToUUIDPtr(order.PaymentID),
		TrackingNumber: order.TrackingNumber,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}
}

// domainOrderItemToAPI converts a domain OrderItem to an API OrderItem
func domainOrderItemToAPI(item *domain.OrderItem) OrderItem {
	subtotal := item.Subtotal()

	return OrderItem{
		ProductId:      stringToUUID(item.ProductID),
		ProductName:    item.ProductName,
		Quantity:       item.Quantity,
		UnitPriceCents: item.UnitPriceCents,
		Subtotal:       subtotal,
	}
}

// createOrderRequestToDTO converts API request to application DTO
func createOrderRequestToDTO(req CreateOrderRequest) application.CreateOrderDTO {
	items := make([]application.OrderItemDTO, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, application.OrderItemDTO{
			ProductID:      uuidToString(item.ProductId),
			ProductName:    item.ProductName,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
		})
	}

	return application.CreateOrderDTO{
		UserID:         uuidToString(req.UserId),
		Items:          items,
		IdempotencyKey: req.IdempotencyKey,
	}
}

// updateStatusRequestToDTO converts API request to application DTO
func updateStatusRequestToDTO(req UpdateStatusRequest) application.UpdateOrderStatusDTO {
	return application.UpdateOrderStatusDTO{
		Status:         domain.OrderStatus(req.Status),
		PaymentID:      uuidPtrToStringPtr(req.PaymentId),
		TrackingNumber: req.TrackingNumber,
		Reason:         req.Reason,
	}
}

// Helper functions for UUID conversions

func stringToUUID(s string) types.UUID {
	u, _ := uuid.Parse(s)
	return types.UUID(u)
}

func uuidToString(u types.UUID) string {
	return uuid.UUID(u).String()
}

func stringPtrToUUIDPtr(s *string) *types.UUID {
	if s == nil {
		return nil
	}
	u, _ := uuid.Parse(*s)
	result := types.UUID(u)
	return &result
}

func uuidPtrToStringPtr(u *types.UUID) *string {
	if u == nil {
		return nil
	}
	str := uuid.UUID(*u).String()
	return &str
}
