package inventory

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rovilay/ecommerce-service/domains/inventory/model"
)

type contextKey string

const InvCTXKey contextKey = "inventory_payload"

func (h *InventoryHandler) MiddlewareValidateInventory(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		inv := &model.InventoryItem{}

		err := inv.FromJSON(r.Body)
		if err != nil {
			h.log.Println("[ERROR] deserializing inventory", err)
			http.Error(w, `{"error": "failed to read payload"}`, http.StatusBadRequest)
			return
		}

		// validate url
		err = inv.Validate()
		if err != nil {
			h.log.Println("[ERROR] validating inventory", err)
			http.Error(
				w, fmt.Sprintf(`{"error": "Error validating inventory: %s"}`, err),
				http.StatusBadRequest,
			)
			return
		}

		// add validated data
		ctx := context.WithValue(r.Context(), InvCTXKey, inv)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
