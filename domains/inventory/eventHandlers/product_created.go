package eventhandlers

import (
	"context"
	"encoding/json"

	"github.com/rovilay/ecommerce-service/common/events"
	eventdatatypes "github.com/rovilay/ecommerce-service/common/events/datatypes"
)

func (h *HandlerClient) ProductCreated(ctx context.Context, e events.EventData) error {
	p := eventdatatypes.Product{}
	if err := json.Unmarshal(e.Data, &p); err != nil {
		h.log.Err(err).Msg("Failed to unmarshal event data")
		return err
	}

	// create product inventory
	_, err := h.repo.CreateInventoryItem(ctx, p.ID, 0)

	return err
}
