package model

type InventoryItem struct {
	ID        int `json:"id"`
	ProductID int `json:"product_id" validate:"required"`
	Quantity  int `json:"quantity"`
}
