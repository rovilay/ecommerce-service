package service

import (
	"context"
	"encoding/json"

	"github.com/rovilay/ecommerce-service/common/events"
	"github.com/rovilay/ecommerce-service/domains/inventory"
	eventhandlers "github.com/rovilay/ecommerce-service/domains/inventory/eventHandlers"
	"github.com/rovilay/ecommerce-service/domains/inventory/model"
	"github.com/rovilay/ecommerce-service/domains/inventory/repository"
	"github.com/rs/zerolog"
)

type InventoryService struct {
	repo repository.InventoryRepository
	rc   *events.RabbitClient
	log  *zerolog.Logger
	hc   *eventhandlers.HandlerClient
}

func NewInventoryService(repo repository.InventoryRepository, rc *events.RabbitClient, l *zerolog.Logger) (*InventoryService, error) {
	logger := l.With().Str("service", "InventoryService").Logger()

	hc := eventhandlers.NewHandlerClient(repo, &logger)

	s := &InventoryService{
		repo: repo,
		rc:   rc,
		log:  &logger,
		hc:   hc,
	}

	return s, nil
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

func (s *InventoryService) Publish(ctx context.Context, topic events.Topic, key events.RoutingKey, e events.EventData) error {
	// create exchange if it doesn't exist
	err := s.rc.CreateExchange(topic, true, false)
	if err != nil {
		s.log.Err(err).Msg("Failed to create exchange")
		return err
	}

	return s.rc.Send(ctx, topic, key, e)
}

func (s *InventoryService) Listen(ctx context.Context, topic events.Topic, key events.RoutingKey) error {
	msgs, err := s.rc.Listen(ctx, events.Product, key, false)
	if err != nil {
		s.log.Err(err).Msg("Failed to create queue binding")
		return err
	}

	go func() {
		for msg := range msgs {
			s.log.Printf("Received an event 🙂: %v", msg.Body)
			e := events.EventData{}
			if err := json.Unmarshal(msg.Body, &e); err != nil {
				s.log.Err(err).Msg("error unmarshalling event")
				continue
			}

			err = s.hc.HandleEvent(ctx, e)
			if err != nil {
				s.log.Err(err).Msgf("Error handling event: %s", msg.MessageId)
				continue
			}

			err = msg.Ack(false)
			if err != nil {
				s.log.Err(err).Msgf("failed to acknowledge message: %s", msg.MessageId)
			}
		}
	}()

	return nil
}
