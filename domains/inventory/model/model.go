package model

import (
	"encoding/json"
	"io"

	"github.com/go-playground/validator/v10"
)

type InventoryItem struct {
	ID        int `json:"id"`
	ProductID int `json:"product_id" db:"product_id" validate:"required"`
	Quantity  int `json:"quantity" db:"quantity" validate:"min=0,required"`
}

func (i *InventoryItem) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(i)
}

func (i *InventoryItem) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(i)
}

func (i *InventoryItem) Validate() error {
	v := validator.New()
	return v.Struct(i)
}
