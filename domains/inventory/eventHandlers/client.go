package eventhandlers

import (
	"context"

	"github.com/rovilay/ecommerce-service/common/events"
	"github.com/rovilay/ecommerce-service/domains/inventory/repository"
	"github.com/rs/zerolog"
)

type HandlerClient struct {
	log  *zerolog.Logger
	repo repository.InventoryRepository
}

func NewHandlerClient(repo repository.InventoryRepository, l *zerolog.Logger) *HandlerClient {
	logger := l.With().Str("inventoryService", "HandlerClient").Logger()

	return &HandlerClient{
		log:  &logger,
		repo: repo,
	}
}

func (h *HandlerClient) HandleEvent(ctx context.Context, event events.EventData) error {
	functionMap, err := h.GetFunctionMap()
	if err != nil {
		h.log.Err(err).Msg("error getting function map")
		return err
	}

	eventFunc, ok := functionMap[string(event.Event)]
	if !ok {
		return nil
	}

	err = eventFunc(ctx, event)
	if err != nil {
		h.log.Err(err).Msg("error handling event")
		return err
	}

	return nil
}

func (h *HandlerClient) GetFunctionMap() (map[string]func(context.Context, events.EventData) error, error) {
	var functionMap = map[string]func(context.Context, events.EventData) error{
		string(events.ProductCreated): h.ProductCreated,
	}

	return functionMap, nil
}
