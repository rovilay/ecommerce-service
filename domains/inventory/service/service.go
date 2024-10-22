package service

import (
	"context"

	"github.com/rovilay/ecommerce-service/domains/inventory"
	"github.com/rovilay/ecommerce-service/domains/inventory/model"
	"github.com/rovilay/ecommerce-service/domains/inventory/repository"
	"github.com/rs/zerolog"
)

type InventoryService struct {
	repo repository.InventoryRepository
	log  *zerolog.Logger
}

func NewInventoryService(repo repository.InventoryRepository, l *zerolog.Logger) *InventoryService {
	logger := l.With().Str("service", "InventoryService").Logger()

	return &InventoryService{
		repo: repo,
		log:  &logger,
	}
}

func (s *InventoryService) CreateInventoryItem(ctx context.Context, productID int, quantity int) (*model.InventoryItem, error) {
	if quantity < 0 {
		return nil, inventory.ErrInvalidQuantity
	}

	return s.repo.CreateInventoryItem(ctx, productID, uint(quantity))
}

func (s *InventoryService) GetInventoryByProductID(ctx context.Context, productID int) (*model.InventoryItem, error) {
	return s.repo.GetInventoryItemByProductID(ctx, productID)
}

func (s *InventoryService) CheckAvailability(ctx context.Context, productID int, quantity uint) (bool, error) {
	inventoryItem, err := s.repo.GetInventoryItemByProductID(ctx, productID)
	if err != nil {
		return false, err
	}

	if inventoryItem.Quantity >= int(quantity) {
		return true, nil
	}

	return false, nil
}

func (s *InventoryService) DecrementInventory(ctx context.Context, productID int, quantity uint) error {
	return s.repo.UpdateInventoryQuantity(ctx, productID, -int(quantity))
}

func (s *InventoryService) IncrementInventory(ctx context.Context, productID int, quantity uint) error {
	return s.repo.UpdateInventoryQuantity(ctx, productID, int(quantity))
}
