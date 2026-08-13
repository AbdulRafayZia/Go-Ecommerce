package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"gocommerce/services/product/internal/application"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ProductHandler implements the ServerInterface
type ProductHandler struct {
	productService  *application.ProductService
	categoryService *application.CategoryService
}

// NewProductHandler creates a new HTTP handler
func NewProductHandler(productService *application.ProductService, categoryService *application.CategoryService) *ProductHandler {
	return &ProductHandler{
		productService:  productService,
		categoryService: categoryService,
	}
}

// HealthCheck implements the health check endpoint
func (h *ProductHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	version := "1.0.0"

	response := HealthResponse{
		Status:    Healthy,
		Timestamp: &now,
		Version:   &version,
	}

	writeJSON(w, http.StatusOK, response)
}

// ListProducts lists products with pagination and filtering
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request, params ListProductsParams) {
	// Build DTO from params
	dto := application.ListProductsDTO{
		Page:     1,
		PageSize: 20,
	}

	if params.Page != nil {
		dto.Page = *params.Page
	}

	if params.PageSize != nil {
		dto.PageSize = *params.PageSize
	}

	if params.CategoryId != nil {
		categoryID := uuidToString(*params.CategoryId)
		dto.CategoryID = &categoryID
	}

	if params.Search != nil {
		dto.Search = params.Search
	}

	if params.MinPrice != nil {
		dto.MinPrice = params.MinPrice
	}

	if params.MaxPrice != nil {
		dto.MaxPrice = params.MaxPrice
	}

	// Call application service
	result, err := h.productService.ListProducts(r.Context(), dto)
	if err != nil {
		handleError(w, err)
		return
	}

	// Map to response
	products := make([]Product, len(result.Products))
	for i, p := range result.Products {
		products[i] = MapProductToResponse(p)
	}

	response := ProductListResponse{
		Data: products,
		Pagination: Pagination{
			Page:       result.Page,
			PageSize:   result.PageSize,
			TotalItems: int(result.TotalItems),
			TotalPages: result.TotalPages,
		},
	}

	writeJSON(w, http.StatusOK, response)
}

// CreateProduct creates a new product
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON request body")
		return
	}

	// Build DTO
	dto := application.CreateProductDTO{
		Name:        req.Name,
		PriceCents:  req.PriceCents,
		Stock:       req.Stock,
		Currency:    "USD", // Default
		Description: "",
	}

	if req.Currency != nil {
		dto.Currency = *req.Currency
	}

	if req.Description != nil {
		dto.Description = *req.Description
	}

	if req.CategoryId != nil {
		categoryID := uuidToString(*req.CategoryId)
		dto.CategoryID = &categoryID
	}

	if req.ImageUrl != nil {
		dto.ImageURL = req.ImageUrl
	}

	// Call application service
	product, err := h.productService.CreateProduct(r.Context(), dto)
	if err != nil {
		handleError(w, err)
		return
	}

	// Map to response
	response := ProductResponse{
		Data: MapProductToResponse(product),
	}

	writeJSON(w, http.StatusCreated, response)
}

// GetProduct retrieves a product by ID
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	productID := uuidToString(id)

	product, err := h.productService.GetProduct(r.Context(), productID)
	if err != nil {
		handleError(w, err)
		return
	}

	response := ProductResponse{
		Data: MapProductToResponse(product),
	}

	writeJSON(w, http.StatusOK, response)
}

// UpdateProduct updates an existing product
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	var req UpdateProductRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON request body")
		return
	}

	productID := uuidToString(id)

	// Build DTO
	dto := application.UpdateProductDTO{
		Name:        req.Name,
		Description: req.Description,
		PriceCents:  req.PriceCents,
		Currency:    req.Currency,
		Stock:       req.Stock,
		Active:      req.Active,
		ImageURL:    req.ImageUrl,
	}

	if req.CategoryId != nil {
		categoryID := uuidToString(*req.CategoryId)
		dto.CategoryID = &categoryID
	}

	// Call application service
	product, err := h.productService.UpdateProduct(r.Context(), productID, dto)
	if err != nil {
		handleError(w, err)
		return
	}

	// Map to response
	response := ProductResponse{
		Data: MapProductToResponse(product),
	}

	writeJSON(w, http.StatusOK, response)
}

// DeleteProduct soft deletes a product
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	productID := uuidToString(id)

	err := h.productService.DeleteProduct(r.Context(), productID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListCategories lists all categories
func (h *ProductHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryService.ListCategories(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	// Map to response
	categoryResponses := make([]Category, len(categories))
	for i, c := range categories {
		categoryResponses[i] = MapCategoryToResponse(c)
	}

	response := CategoryListResponse{
		Data: categoryResponses,
	}

	writeJSON(w, http.StatusOK, response)
}

// CreateCategory creates a new category
func (h *ProductHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON request body")
		return
	}

	// Build DTO
	dto := application.CreateCategoryDTO{
		Name:        req.Name,
		Description: req.Description,
	}

	if req.ParentId != nil {
		parentID := uuidToString(*req.ParentId)
		dto.ParentID = &parentID
	}

	// Call application service
	category, err := h.categoryService.CreateCategory(r.Context(), dto)
	if err != nil {
		handleError(w, err)
		return
	}

	// Map to response
	response := CategoryResponse{
		Data: MapCategoryToResponse(category),
	}

	writeJSON(w, http.StatusCreated, response)
}
