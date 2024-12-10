package events

import (
	"encoding/json"
	"time"
)

type EventData struct {
	Event RoutingKey      `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type ProductCreatedEvent struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float32   `json:"price"`
	SKU         string    `json:"sku" db:"sku"`
	ImageURL    string    `json:"image_url"`
	CategoryID  int       `json:"category_id"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	DeletedAt   string    `json:"deleted_at,omitempty"`
}
