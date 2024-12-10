package eventdatatypes

import "time"

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
