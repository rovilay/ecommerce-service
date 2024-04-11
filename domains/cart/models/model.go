package models

import (
	"encoding/json"
	"io"
	"time"
)

type Cart struct {
	ID        int        `json:"id"`
	UserID    string     `json:"user_id"`
	CartItems []CartItem `json:"cart_items"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CartItem struct {
	ID        int `json:"id"`
	CartID    int `json:"cart_id"`
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

func (c *Cart) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(c)
}

func (c *Cart) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(c)
}

func (i *CartItem) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(i)
}

func (i *CartItem) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(i)
}
