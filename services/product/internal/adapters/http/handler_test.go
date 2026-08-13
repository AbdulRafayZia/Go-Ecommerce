package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gocommerce/services/product/internal/application"
	"gocommerce/services/product/internal/application/mocks"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func setupTestHandler() *ProductHandler {
	productRepo := mocks.NewMockProductRepository()
	categoryRepo := mocks.NewMockCategoryRepository()

	productService := application.NewProductService(productRepo, categoryRepo)
	categoryService := application.NewCategoryService(categoryRepo)

	return NewProductHandler(productService, categoryService)
}

func TestHandler_HealthCheck(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HealthCheck() status = %v, want %v", w.Code, http.StatusOK)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != Healthy {
		t.Errorf("HealthResponse.Status = %v, want %v", response.Status, Healthy)
	}
}

func TestHandler_CreateProduct(t *testing.T) {
	handler := setupTestHandler()

	tests := []struct {
		name       string
		body       CreateProductRequest
		wantStatus int
	}{
		{
			name: "valid product",
			body: CreateProductRequest{
				Name:        "Test Product",
				Description: stringPtr("Test Description"),
				PriceCents:  1000,
				Stock:       10,
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "product with zero price",
			body: CreateProductRequest{
				Name:        "Free Product",
				Description: stringPtr("Free item"),
				PriceCents:  0,
				Stock:       5,
			},
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/v1/products", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateProduct(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateProduct() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if w.Code == http.StatusCreated {
				var response ProductResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response.Data.Name != tt.body.Name {
					t.Errorf("Product.Name = %v, want %v", response.Data.Name, tt.body.Name)
				}

				if response.Data.PriceCents != tt.body.PriceCents {
					t.Errorf("Product.PriceCents = %v, want %v", response.Data.PriceCents, tt.body.PriceCents)
				}
			}
		})
	}
}

func TestHandler_CreateProduct_InvalidRequest(t *testing.T) {
	handler := setupTestHandler()

	tests := []struct {
		name string
		body string
	}{
		{
			name: "invalid JSON",
			body: `{invalid json}`,
		},
		{
			name: "empty body",
			body: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/products", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateProduct(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("CreateProduct() status = %v, want %v", w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestHandler_GetProduct(t *testing.T) {
	handler := setupTestHandler()

	// First create a product
	createReq := CreateProductRequest{
		Name:        "Test Product",
		Description: stringPtr("Test Description"),
		PriceCents:  1000,
		Stock:       10,
	}
	bodyBytes, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/v1/products", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateProduct(w, req)

	var createResponse ProductResponse
	json.NewDecoder(w.Body).Decode(&createResponse)
	productID := createResponse.Data.Id

	t.Run("get existing product", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/products/"+productID.String(), nil)
		w := httptest.NewRecorder()

		handler.GetProduct(w, req, productID)

		if w.Code != http.StatusOK {
			t.Errorf("GetProduct() status = %v, want %v", w.Code, http.StatusOK)
		}

		var response ProductResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Data.Id != productID {
			t.Errorf("Product.ID = %v, want %v", response.Data.Id, productID)
		}
	})

	t.Run("get non-existent product", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/v1/products/"+nonExistentID.String(), nil)
		w := httptest.NewRecorder()

		handler.GetProduct(w, req, nonExistentID)

		if w.Code != http.StatusNotFound {
			t.Errorf("GetProduct() status = %v, want %v", w.Code, http.StatusNotFound)
		}
	})
}

func TestHandler_UpdateProduct(t *testing.T) {
	handler := setupTestHandler()

	// First create a product
	createReq := CreateProductRequest{
		Name:        "Test Product",
		Description: stringPtr("Test Description"),
		PriceCents:  1000,
		Stock:       10,
	}
	bodyBytes, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/v1/products", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateProduct(w, req)

	var createResponse ProductResponse
	json.NewDecoder(w.Body).Decode(&createResponse)
	productID := createResponse.Data.Id

	t.Run("update product name", func(t *testing.T) {
		newName := "Updated Product"
		updateReq := UpdateProductRequest{
			Name: &newName,
		}

		bodyBytes, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(http.MethodPut, "/v1/products/"+productID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.UpdateProduct(w, req, productID)

		if w.Code != http.StatusOK {
			t.Errorf("UpdateProduct() status = %v, want %v", w.Code, http.StatusOK)
		}

		var response ProductResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Data.Name != newName {
			t.Errorf("Product.Name = %v, want %v", response.Data.Name, newName)
		}
	})

	t.Run("update product price", func(t *testing.T) {
		newPrice := 2000
		updateReq := UpdateProductRequest{
			PriceCents: &newPrice,
		}

		bodyBytes, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(http.MethodPut, "/v1/products/"+productID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.UpdateProduct(w, req, productID)

		if w.Code != http.StatusOK {
			t.Errorf("UpdateProduct() status = %v, want %v", w.Code, http.StatusOK)
		}

		var response ProductResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Data.PriceCents != newPrice {
			t.Errorf("Product.PriceCents = %v, want %v", response.Data.PriceCents, newPrice)
		}
	})
}

func TestHandler_DeleteProduct(t *testing.T) {
	handler := setupTestHandler()

	// First create a product
	createReq := CreateProductRequest{
		Name:        "Test Product",
		Description: stringPtr("Test Description"),
		PriceCents:  1000,
		Stock:       10,
	}
	bodyBytes, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/v1/products", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateProduct(w, req)

	var createResponse ProductResponse
	json.NewDecoder(w.Body).Decode(&createResponse)
	productID := createResponse.Data.Id

	t.Run("delete existing product", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/v1/products/"+productID.String(), nil)
		w := httptest.NewRecorder()

		handler.DeleteProduct(w, req, productID)

		if w.Code != http.StatusNoContent {
			t.Errorf("DeleteProduct() status = %v, want %v", w.Code, http.StatusNoContent)
		}
	})

	t.Run("delete non-existent product", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodDelete, "/v1/products/"+nonExistentID.String(), nil)
		w := httptest.NewRecorder()

		handler.DeleteProduct(w, req, nonExistentID)

		if w.Code != http.StatusNotFound {
			t.Errorf("DeleteProduct() status = %v, want %v", w.Code, http.StatusNotFound)
		}
	})
}

func TestHandler_ListProducts(t *testing.T) {
	handler := setupTestHandler()

	// Create some test products
	for i := 1; i <= 3; i++ {
		createReq := CreateProductRequest{
			Name:        "Product " + string(rune(i+'0')),
			Description: stringPtr("Test Description"),
			PriceCents:  1000 * i,
			Stock:       10,
		}
		bodyBytes, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/v1/products", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateProduct(w, req)
	}

	t.Run("list all products", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/products", nil)
		w := httptest.NewRecorder()

		handler.ListProducts(w, req, ListProductsParams{})

		if w.Code != http.StatusOK {
			t.Errorf("ListProducts() status = %v, want %v", w.Code, http.StatusOK)
		}

		var response ProductListResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(response.Data) != 3 {
			t.Errorf("len(Products) = %v, want 3", len(response.Data))
		}
	})
}

func TestHandler_CreateCategory(t *testing.T) {
	handler := setupTestHandler()

	t.Run("create valid category", func(t *testing.T) {
		createReq := CreateCategoryRequest{
			Name:        "Electronics",
			Description: stringPtr("Electronic devices"),
		}

		bodyBytes, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/v1/categories", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateCategory(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("CreateCategory() status = %v, want %v", w.Code, http.StatusCreated)
		}

		var response CategoryResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Data.Name != createReq.Name {
			t.Errorf("Category.Name = %v, want %v", response.Data.Name, createReq.Name)
		}
	})
}

func TestHandler_ListCategories(t *testing.T) {
	handler := setupTestHandler()

	// Create some test categories
	for i := 1; i <= 2; i++ {
		createReq := CreateCategoryRequest{
			Name:        "Category " + string(rune(i+'0')),
			Description: stringPtr("Test Description"),
		}
		bodyBytes, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/v1/categories", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateCategory(w, req)
	}

	t.Run("list all categories", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/categories", nil)
		w := httptest.NewRecorder()

		handler.ListCategories(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("ListCategories() status = %v, want %v", w.Code, http.StatusOK)
		}

		var response CategoryListResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(response.Data) != 2 {
			t.Errorf("len(Categories) = %v, want 2", len(response.Data))
		}
	})
}

func TestHandler_IntegrationWithRouter(t *testing.T) {
	handler := setupTestHandler()

	// Create a router and register the handler
	r := chi.NewRouter()
	HandlerFromMux(handler, r)

	// Test creating a product through the router
	t.Run("create and retrieve product via router", func(t *testing.T) {
		createReq := CreateProductRequest{
			Name:        "Integration Test Product",
			Description: stringPtr("Integration test"),
			PriceCents:  1500,
			Stock:       20,
		}

		bodyBytes, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/v1/products", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Create product via router failed: status = %v", w.Code)
		}

		var createResponse ProductResponse
		json.NewDecoder(w.Body).Decode(&createResponse)

		// Now retrieve the product
		req = httptest.NewRequest(http.MethodGet, "/v1/products/"+createResponse.Data.Id.String(), nil)
		w = httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get product via router failed: status = %v", w.Code)
		}
	})
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
