package product

import (
	"context"
)

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Price       float32 `json:"price" validate:"gt=0"`
	SKU         string  `json:"sku" validate:"required"`
	ImageURL    string  `json:"image_url" validate:"required"`
	CategoryID  int     `json:"category_id" validate:"required"`
	CreatedAt   string  `json:"created_at,omitempty"`
	UpdatedAt   string  `json:"updated_at,omitempty"`
	DeletedAt   string  `json:"deleted_at,omitempty"`
}

type ProductsWithPagination struct {
	products []*Product
	limit    int
	offset   int
	total    int
}

type Operations interface {
	GetProduct(tx context.Context, id int) (*Product, error)
	CreateProduct(tx context.Context, data *Product) (*Product, error)
	ListProducts(tx context.Context, limit int, offset int) (*ProductsWithPagination, error)
	UpateProduct(tx context.Context, id int, data *Product) (*Product, error)
	DeleteProduct(tx context.Context, id int) error
	SearchProductsByName(ctx context.Context, searchTerm string) ([]*Product, error)
}

type Repository interface {
	GetProductByID(ctx context.Context, id int) (*Product, error)
	GetAllProducts(ctx context.Context, limit int, offset int) ([]*Product, error)
	CreateProduct(ctx context.Context, p *Product) (*Product, error)
	UpdateProduct(ctx context.Context, p *Product) (*Product, error)
	GetProductsByCategory(ctx context.Context, categoryID int) ([]*Product, error)
	DeleteProduct(ctx context.Context, id int) error
	CountProducts(ctx context.Context) (int, error)
	SearchProductsByName(ctx context.Context, searchTerm string) ([]*Product, error)
}
