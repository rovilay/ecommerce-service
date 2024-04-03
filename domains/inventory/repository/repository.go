package repository

import (
	"context"

	"github.com/rovilay/ecommerce-service/domains/inventory/model"
)

type InventoryRepository interface {
	CreateInventoryItem(ctx context.Context, productID int, quantity uint) (*model.InventoryItem, error)
	GetInventoryItemByProductID(ctx context.Context, productID int) (*model.InventoryItem, error)
	UpdateInventoryQuantity(ctx context.Context, productID, quantityDelta int) error
}
