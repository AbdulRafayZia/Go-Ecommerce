package api

import (
	"gocommerce/services/cart/internal/domain"
)

// domainCartToAPI converts a domain Cart to an API Cart response
func domainCartToAPI(cart *domain.Cart) *Cart {
	apiItems := make([]CartItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		apiItems = append(apiItems, domainCartItemToAPI(item))
	}

	itemCount := cart.ItemCount()
	totalQty := cart.TotalQuantity()
	totalPrice := cart.TotalPrice()

	return &Cart{
		UserId:        cart.UserID,
		Items:         apiItems,
		ItemCount:     &itemCount,
		TotalQuantity: &totalQty,
		TotalPrice:    totalPrice,
		CreatedAt:     cart.CreatedAt,
		UpdatedAt:     cart.UpdatedAt,
	}
}

// domainCartItemToAPI converts a domain CartItem to an API CartItem
func domainCartItemToAPI(item *domain.CartItem) CartItem {
	subtotal := item.Subtotal()

	return CartItem{
		ProductId:  item.ProductID,
		Name:       item.Name,
		PriceCents: item.PriceCents,
		Quantity:   item.Quantity,
		Subtotal:   subtotal,
		ImageUrl:   item.ImageURL,
	}
}
