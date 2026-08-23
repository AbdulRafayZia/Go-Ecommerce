package http

import (
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"gocommerce/services/payment/internal/domain"
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

// domainPaymentToResponse converts domain Payment to PaymentResponse
func domainPaymentToResponse(payment *domain.Payment) PaymentResponse {
	resp := PaymentResponse{
		Id:            stringToUUID(payment.ID),
		OrderId:       stringToUUID(payment.OrderID),
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		Status:        PaymentResponseStatus(payment.Status),
		PaymentMethod: (*string)(strPtr(string(payment.PaymentMethod))),
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}

	if payment.ProviderPaymentID != "" {
		resp.ProviderPaymentId = strPtr(payment.ProviderPaymentID)
	}

	if payment.FailureReason != "" {
		resp.FailureReason = strPtr(payment.FailureReason)
	}

	return resp
}

// strPtr returns a pointer to a string
func strPtr(s string) *string {
	return &s
}
