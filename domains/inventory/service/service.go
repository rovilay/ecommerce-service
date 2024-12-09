package service

import (
	"context"

	"github.com/rovilay/ecommerce-service/common/events"
	"github.com/rovilay/ecommerce-service/domains/inventory"
	"github.com/rovilay/ecommerce-service/domains/inventory/model"
	"github.com/rovilay/ecommerce-service/domains/inventory/repository"
	"github.com/rs/zerolog"
)

type InventoryService struct {
	repo      repository.InventoryRepository
	msgBroker *events.RabbitClient
	log       *zerolog.Logger
}

func NewInventoryService(repo repository.InventoryRepository, b *events.RabbitClient, l *zerolog.Logger) (*InventoryService, error) {
	logger := l.With().Str("service", "InventoryService").Logger()

	s := &InventoryService{
		repo:      repo,
		msgBroker: b,
		log:       &logger,
	}

	err := s.setupListeners()
	if err != nil {
		return nil, err
	}

	return s, err
}

func (s *InventoryService) setupListeners() error {
	productCreatedMsgs, err := s.msgBroker.Consume(events.ProductCreated, events.Product, false)
	if err != nil {
		s.log.Err(err).Msg("Failed to register a consumer")
		return err
	}

	go func() {
		for msg := range productCreatedMsgs {
			// e := events.EventData{}
			// json.Unmarshal(msg, e)
			// s.log.Printf("Received a event: %s, data: %+v\n")
			s.log.Printf(" [x] %s", msg.Body)
		}
	}()

	return nil
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
