package api

import (
	"gocommerce/services/product/internal/domain"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// MapProductToResponse converts domain.Product to API Product
func MapProductToResponse(p *domain.Product) Product {
	product := Product{
		Id:          mustParseUUID(p.ID),
		Name:        p.Name,
		PriceCents:  p.PriceCents,
		Currency:    p.Currency,
		Stock:       p.Stock,
		Active:      p.Active,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	if p.Description != "" {
		product.Description = &p.Description
	}

	if p.CategoryID != nil {
		categoryUUID := mustParseUUID(*p.CategoryID)
		product.CategoryId = &categoryUUID
	}

	if p.ImageURL != nil {
		product.ImageUrl = p.ImageURL
	}

	return product
}

// MapCategoryToResponse converts domain.Category to API Category
func MapCategoryToResponse(c *domain.Category) Category {
	category := Category{
		Id:        mustParseUUID(c.ID),
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
	}

	if c.Description != nil {
		category.Description = c.Description
	}

	if c.ParentID != nil {
		parentUUID := mustParseUUID(*c.ParentID)
		category.ParentId = &parentUUID
	}

	return category
}

// mustParseUUID parses a UUID string and panics on error
// This should only be used with UUIDs from the database that we know are valid
func mustParseUUID(s string) openapi_types.UUID {
	u, err := uuid.Parse(s)
	if err != nil {
		panic("invalid UUID from database: " + s)
	}
	return openapi_types.UUID(u)
}

// uuidToStringPtr converts UUID pointer to string pointer
func uuidToStringPtr(u *openapi_types.UUID) *string {
	if u == nil {
		return nil
	}
	s := u.String()
	return &s
}

// uuidToString converts UUID to string
func uuidToString(u openapi_types.UUID) string {
	return uuid.UUID(u).String()
}
