package http

import (
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"gocommerce/services/inventory/internal/domain"
)

// uuidToString converts an openapi UUID to string
func uuidToString(u openapi_types.UUID) string {
	return uuid.UUID(u).String()
}

// stringToUUID converts a string to openapi UUID
func stringToUUID(s string) openapi_types.UUID {
	u, _ := uuid.Parse(s)
	return openapi_types.UUID(u)
}

// domainStockToResponse converts domain Stock to StockResponse
func domainStockToResponse(stock *domain.Stock) StockResponse {
	total := stock.TotalStock()

	return StockResponse{
		ProductId: stringToUUID(stock.ProductID),
		Available: stock.Available,
		Reserved:  stock.Reserved,
		Total:     total,
		UpdatedAt: stock.UpdatedAt,
	}
}

// domainReservationToItem converts domain Reservation to ReservationItem
func domainReservationToItem(reservation *domain.Reservation) ReservationItem {
	// Convert domain status to API status string
	var status string
	switch reservation.Status {
	case domain.ReservationStatusPending:
		status = "pending"
	case domain.ReservationStatusFulfilled:
		status = "fulfilled"
	case domain.ReservationStatusCancelled:
		status = "cancelled"
	case domain.ReservationStatusExpired:
		status = "expired"
	default:
		status = "pending"
	}

	return ReservationItem{
		Id:        stringToUUID(reservation.ID),
		OrderId:   stringToUUID(reservation.OrderID),
		ProductId: stringToUUID(reservation.ProductID),
		Quantity:  reservation.Quantity,
		Status:    ReservationItemStatus(status),
		CreatedAt: reservation.CreatedAt,
		UpdatedAt: reservation.UpdatedAt,
	}
}
