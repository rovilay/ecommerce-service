package product

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/go-playground/validator/v10"
)

type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	Price       float32   `json:"price" db:"price" validate:"gt=0"`
	SKU         string    `json:"sku" db:"sku" validate:"required,len=5"`
	ImageURL    string    `json:"image_url" db:"image_url"`
	CategoryID  int       `json:"category_id" db:"category_id" validate:"required"`
	CreatedAt   time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" db:"updated_at"`
	DeletedAt   string    `json:"deleted_at,omitempty" db:"deleted_at"`
}

type Category struct {
	ID        int       `json:"id"`
	Name      string    `json:"name" validate:"required,min=3"`
	CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

type PaginationResult[T any] struct {
	Items  []T `json:"items"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
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

	GetCategoryByID(ctx context.Context, id int) (*Category, error)
	GetAllCategories(ctx context.Context, limit, offset int) ([]*Category, error)
	CreateCategory(ctx context.Context, name string) (*Category, error)
	UpdateCategory(ctx context.Context, id int, name string) (*Category, error)
	SearchCategoriesByName(ctx context.Context, searchTerm string) ([]*Category, error)
	CountCategories(ctx context.Context) (int, error)
}

func (p *Product) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(p)
}

func (p *Product) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(p)
}

func (p *Product) Validate() error {
	v := validator.New()
	return v.Struct(p)
}

func (c *Category) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(c)
}

func (c *Category) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(c)
}

func (c *Category) Validate() error {
	v := validator.New()
	return v.Struct(c)
}
