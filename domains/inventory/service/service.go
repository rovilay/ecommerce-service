package service

import (
	"context"

	"github.com/rovilay/ecommerce-service/domains/inventory"
	"github.com/rovilay/ecommerce-service/domains/inventory/model"
	"github.com/rovilay/ecommerce-service/domains/inventory/repository"
)

type InventoryService struct {
	repo repository.InventoryRepository
}

func NewInventoryService(repo *repository.InventoryRepository) *InventoryService {
	return &InventoryService{
		repo: *repo,
	}
}

func (s *InventoryService) CreateInventory(ctx context.Context, productID int, quantity uint) (*model.InventoryItem, error) {
	return s.repo.CreateInventoryItem(ctx, productID, quantity)
}

func (s *InventoryService) GetInventory(ctx context.Context, productID int) (*model.InventoryItem, error) {
	return s.repo.GetInventoryItemByProductID(ctx, productID)
}

func (s *InventoryService) CheckAvailability(ctx context.Context, productID int, quantity int) (bool, error) {
	inventoryItem, err := s.repo.GetInventoryItemByProductID(ctx, productID)
	if err != nil {
		return false, err
	}

	if inventoryItem.Quantity >= quantity {
		return true, nil
	}

	return false, inventory.ErrInsufficientStock
}

func (s *InventoryService) DecrementInventory(ctx context.Context, productID int, quantity int) error {
	err := s.repo.UpdateInventoryQuantity(ctx, productID, -quantity)

	return err
}

func (s *InventoryService) IncrementInventory(ctx context.Context, productID int, quantity int) error {
	err := s.repo.UpdateInventoryQuantity(ctx, productID, quantity)

	return err
}
