package application

import (
	"context"
	"testing"

	"gocommerce/services/product/internal/application/mocks"
	"gocommerce/services/product/internal/domain"
)

func TestProductService_CreateProduct(t *testing.T) {
	productRepo := mocks.NewMockProductRepository()
	categoryRepo := mocks.NewMockCategoryRepository()
	service := NewProductService(productRepo, categoryRepo)

	ctx := context.Background()

	t.Run("create valid product", func(t *testing.T) {
		dto := CreateProductDTO{
			Name:        "Test Product",
			Description: "Test Description",
			PriceCents:  1000,
			Currency:    "USD",
			Stock:       10,
		}

		product, err := service.CreateProduct(ctx, dto)
		if err != nil {
			t.Fatalf("CreateProduct() error = %v", err)
		}

		if product.Name != dto.Name {
			t.Errorf("Product.Name = %v, want %v", product.Name, dto.Name)
		}

		if product.PriceCents != dto.PriceCents {
			t.Errorf("Product.PriceCents = %v, want %v", product.PriceCents, dto.PriceCents)
		}
	})

	t.Run("create product with invalid category", func(t *testing.T) {
		invalidCategoryID := "non-existent-category"
		dto := CreateProductDTO{
			Name:        "Test Product 2",
			Description: "Test Description",
			PriceCents:  1000,
			Currency:    "USD",
			CategoryID:  &invalidCategoryID,
			Stock:       10,
		}

		_, err := service.CreateProduct(ctx, dto)
		if err == nil {
			t.Error("CreateProduct() should fail with invalid category")
		}
	})

	t.Run("create product with duplicate name", func(t *testing.T) {
		dto := CreateProductDTO{
			Name:        "Test Product",
			Description: "Test Description",
			PriceCents:  1000,
			Currency:    "USD",
			Stock:       10,
		}

		_, err := service.CreateProduct(ctx, dto)
		if err != domain.ErrProductAlreadyExists {
			t.Errorf("CreateProduct() error = %v, want %v", err, domain.ErrProductAlreadyExists)
		}
	})

	t.Run("create product with valid category", func(t *testing.T) {
		// First create a category
		category, _ := domain.NewCategory("Electronics", nil, nil)
		category.ID = "cat-123"
		categoryRepo.Create(ctx, category)

		dto := CreateProductDTO{
			Name:        "Product with Category",
			Description: "Test Description",
			PriceCents:  1000,
			Currency:    "USD",
			CategoryID:  &category.ID,
			Stock:       10,
		}

		product, err := service.CreateProduct(ctx, dto)
		if err != nil {
			t.Fatalf("CreateProduct() error = %v", err)
		}

		if product.CategoryID == nil {
			t.Error("Product.CategoryID should not be nil")
		} else if *product.CategoryID != category.ID {
			t.Errorf("Product.CategoryID = %v, want %v", *product.CategoryID, category.ID)
		}
	})

	t.Run("create product with invalid data", func(t *testing.T) {
		dto := CreateProductDTO{
			Name:        "",
			Description: "Test Description",
			PriceCents:  1000,
			Currency:    "USD",
			Stock:       10,
		}

		_, err := service.CreateProduct(ctx, dto)
		if err == nil {
			t.Error("CreateProduct() should fail with empty name")
		}
	})
}

func TestProductService_GetProduct(t *testing.T) {
	productRepo := mocks.NewMockProductRepository()
	categoryRepo := mocks.NewMockCategoryRepository()
	service := NewProductService(productRepo, categoryRepo)

	ctx := context.Background()

	// Create a test product
	dto := CreateProductDTO{
		Name:        "Test Product",
		Description: "Test Description",
		PriceCents:  1000,
		Currency:    "USD",
		Stock:       10,
	}
	createdProduct, _ := service.CreateProduct(ctx, dto)

	t.Run("get existing product", func(t *testing.T) {
		product, err := service.GetProduct(ctx, createdProduct.ID)
		if err != nil {
			t.Fatalf("GetProduct() error = %v", err)
		}

		if product.ID != createdProduct.ID {
			t.Errorf("Product.ID = %v, want %v", product.ID, createdProduct.ID)
		}
	})

	t.Run("get non-existent product", func(t *testing.T) {
		_, err := service.GetProduct(ctx, "non-existent-id")
		if err != domain.ErrProductNotFound {
			t.Errorf("GetProduct() error = %v, want %v", err, domain.ErrProductNotFound)
		}
	})
}

func TestProductService_UpdateProduct(t *testing.T) {
	productRepo := mocks.NewMockProductRepository()
	categoryRepo := mocks.NewMockCategoryRepository()
	service := NewProductService(productRepo, categoryRepo)

	ctx := context.Background()

	// Create a test product
	dto := CreateProductDTO{
		Name:        "Test Product",
		Description: "Test Description",
		PriceCents:  1000,
		Currency:    "USD",
		Stock:       10,
	}
	createdProduct, _ := service.CreateProduct(ctx, dto)

	t.Run("update product name", func(t *testing.T) {
		newName := "Updated Product"
		updateDTO := UpdateProductDTO{
			Name: &newName,
		}

		product, err := service.UpdateProduct(ctx, createdProduct.ID, updateDTO)
		if err != nil {
			t.Fatalf("UpdateProduct() error = %v", err)
		}

		if product.Name != newName {
			t.Errorf("Product.Name = %v, want %v", product.Name, newName)
		}
	})

	t.Run("update product price", func(t *testing.T) {
		newPrice := 2000
		updateDTO := UpdateProductDTO{
			PriceCents: &newPrice,
		}

		product, err := service.UpdateProduct(ctx, createdProduct.ID, updateDTO)
		if err != nil {
			t.Fatalf("UpdateProduct() error = %v", err)
		}

		if product.PriceCents != newPrice {
			t.Errorf("Product.PriceCents = %v, want %v", product.PriceCents, newPrice)
		}
	})

	t.Run("update product stock", func(t *testing.T) {
		newStock := 20
		updateDTO := UpdateProductDTO{
			Stock: &newStock,
		}

		product, err := service.UpdateProduct(ctx, createdProduct.ID, updateDTO)
		if err != nil {
			t.Fatalf("UpdateProduct() error = %v", err)
		}

		if product.Stock != newStock {
			t.Errorf("Product.Stock = %v, want %v", product.Stock, newStock)
		}
	})

	t.Run("deactivate product", func(t *testing.T) {
		active := false
		updateDTO := UpdateProductDTO{
			Active: &active,
		}

		product, err := service.UpdateProduct(ctx, createdProduct.ID, updateDTO)
		if err != nil {
			t.Fatalf("UpdateProduct() error = %v", err)
		}

		if product.Active {
			t.Error("Product.Active should be false")
		}
	})

	t.Run("update non-existent product", func(t *testing.T) {
		newName := "Updated Product"
		updateDTO := UpdateProductDTO{
			Name: &newName,
		}

		_, err := service.UpdateProduct(ctx, "non-existent-id", updateDTO)
		if err != domain.ErrProductNotFound {
			t.Errorf("UpdateProduct() error = %v, want %v", err, domain.ErrProductNotFound)
		}
	})
}

func TestProductService_DeleteProduct(t *testing.T) {
	productRepo := mocks.NewMockProductRepository()
	categoryRepo := mocks.NewMockCategoryRepository()
	service := NewProductService(productRepo, categoryRepo)

	ctx := context.Background()

	// Create a test product
	dto := CreateProductDTO{
		Name:        "Test Product",
		Description: "Test Description",
		PriceCents:  1000,
		Currency:    "USD",
		Stock:       10,
	}
	createdProduct, _ := service.CreateProduct(ctx, dto)

	t.Run("delete existing product", func(t *testing.T) {
		err := service.DeleteProduct(ctx, createdProduct.ID)
		if err != nil {
			t.Fatalf("DeleteProduct() error = %v", err)
		}

		// Verify product is deactivated (soft delete)
		product, _ := service.GetProduct(ctx, createdProduct.ID)
		if product.Active {
			t.Error("Product should be inactive after deletion")
		}
	})

	t.Run("delete non-existent product", func(t *testing.T) {
		err := service.DeleteProduct(ctx, "non-existent-id")
		if err != domain.ErrProductNotFound {
			t.Errorf("DeleteProduct() error = %v, want %v", err, domain.ErrProductNotFound)
		}
	})
}

func TestProductService_ListProducts(t *testing.T) {
	productRepo := mocks.NewMockProductRepository()
	categoryRepo := mocks.NewMockCategoryRepository()
	service := NewProductService(productRepo, categoryRepo)

	ctx := context.Background()

	// Create test products
	for i := 1; i <= 5; i++ {
		dto := CreateProductDTO{
			Name:        "Product " + string(rune(i+'0')),
			Description: "Test Description",
			PriceCents:  1000 * i,
			Currency:    "USD",
			Stock:       10,
		}
		service.CreateProduct(ctx, dto)
	}

	t.Run("list products with default pagination", func(t *testing.T) {
		dto := ListProductsDTO{}

		result, err := service.ListProducts(ctx, dto)
		if err != nil {
			t.Fatalf("ListProducts() error = %v", err)
		}

		if len(result.Products) != 5 {
			t.Errorf("len(Products) = %v, want 5", len(result.Products))
		}

		if result.TotalItems != 5 {
			t.Errorf("TotalItems = %v, want 5", result.TotalItems)
		}
	})

	t.Run("list products with page size", func(t *testing.T) {
		pageSize := 2
		dto := ListProductsDTO{
			PageSize: pageSize,
		}

		result, err := service.ListProducts(ctx, dto)
		if err != nil {
			t.Fatalf("ListProducts() error = %v", err)
		}

		if result.PageSize != pageSize {
			t.Errorf("PageSize = %v, want %v", result.PageSize, pageSize)
		}
	})
}
