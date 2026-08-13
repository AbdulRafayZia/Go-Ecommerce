package domain

import (
	"testing"
)

func TestNewCategory(t *testing.T) {
	tests := []struct {
		name         string
		categoryName string
		description  *string
		parentID     *string
		wantErr      error
	}{
		{
			name:         "valid category",
			categoryName: "Electronics",
			description:  stringPtr("Electronic devices"),
			parentID:     nil,
			wantErr:      nil,
		},
		{
			name:         "valid category with parent",
			categoryName: "Smartphones",
			description:  stringPtr("Mobile phones"),
			parentID:     stringPtr("parent-123"),
			wantErr:      nil,
		},
		{
			name:         "valid category without description",
			categoryName: "Books",
			description:  nil,
			parentID:     nil,
			wantErr:      nil,
		},
		{
			name:         "empty name",
			categoryName: "",
			description:  stringPtr("Description"),
			parentID:     nil,
			wantErr:      ErrEmptyCategoryName,
		},
		{
			name:         "whitespace name",
			categoryName: "   ",
			description:  stringPtr("Description"),
			parentID:     nil,
			wantErr:      ErrEmptyCategoryName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category, err := NewCategory(tt.categoryName, tt.description, tt.parentID)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("NewCategory() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("NewCategory() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("NewCategory() unexpected error = %v", err)
				return
			}

			if category == nil {
				t.Error("NewCategory() returned nil category")
				return
			}

			if category.Name != tt.categoryName {
				t.Errorf("Category.Name = %v, want %v", category.Name, tt.categoryName)
			}

			if tt.description != nil {
				if category.Description == nil {
					t.Error("Category.Description should not be nil")
				} else if *category.Description != *tt.description {
					t.Errorf("Category.Description = %v, want %v", *category.Description, *tt.description)
				}
			} else if category.Description != nil {
				t.Error("Category.Description should be nil")
			}

			if tt.parentID != nil {
				if category.ParentID == nil {
					t.Error("Category.ParentID should not be nil")
				} else if *category.ParentID != *tt.parentID {
					t.Errorf("Category.ParentID = %v, want %v", *category.ParentID, *tt.parentID)
				}
			} else if category.ParentID != nil {
				t.Error("Category.ParentID should be nil")
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
