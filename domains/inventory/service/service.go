package service

import (
	"context"

	"github.com/rovilay/ecommerce-service/domains/inventory"
	"github.com/rovilay/ecommerce-service/domains/inventory/model"
	"github.com/rovilay/ecommerce-service/domains/inventory/repository"
	"github.com/rs/zerolog"
)

type ProductServiceClient interface {
	CheckProductExists(ctx context.Context, productID int) (bool, error)
}

type InventoryService struct {
	repo                 repository.InventoryRepository
	productServiceClient ProductServiceClient
	log                  *zerolog.Logger
}

func NewInventoryService(repo *repository.InventoryRepository, psc *ProductServiceClient, l *zerolog.Logger) *InventoryService {
	logger := l.With().Str("repository", "postgresInventoryRepository").Logger()

	return &InventoryService{
		repo:                 *repo,
		productServiceClient: *psc,
		log:                  &logger,
	}
}

func (s *InventoryService) CreateInventoryItem(ctx context.Context, productID int, quantity int) (*model.InventoryItem, error) {
	log := s.log.With().Str("method", "CreateInventoryItem").Logger()

	if quantity < 0 {
		return nil, inventory.ErrInvalidQuantity
	}

	productExists, err := s.verifyProductExists(ctx, productID)
	if !productExists || err != nil {
		if err != nil {
			log.Err(err)
		}
		return nil, inventory.ErrInvalidProduct
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

	return false, inventory.ErrInsufficientStock
}

func (s *InventoryService) DecrementInventory(ctx context.Context, productID int, quantity uint) error {
	err := s.repo.UpdateInventoryQuantity(ctx, productID, -int(quantity))

	return err
}

func (s *InventoryService) IncrementInventory(ctx context.Context, productID int, quantity uint) error {
	err := s.repo.UpdateInventoryQuantity(ctx, productID, int(quantity))

	return err
}

func (s *InventoryService) verifyProductExists(ctx context.Context, productID int) (bool, error) {
	return s.productServiceClient.CheckProductExists(ctx, productID)
}
