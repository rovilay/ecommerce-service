package order

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rovilay/ecommerce-service/common/utils"
	"github.com/rovilay/ecommerce-service/domains/order/models"
)

type contextKey string

const OrderCTXKey contextKey = "cart_item_payload"
const AuthCTXKey contextKey = "auth_token"

func (h *OrderHandler) MiddlewareValidateOrderItems(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		order := &models.Order{}

		if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
			h.log.Println("[ERROR] deserializing order items", err)
			http.Error(w, `{"error": "failed to read payload"}`, http.StatusBadRequest)
			return
		}

		// validate url
		err := order.Validate()
		if err != nil {
			h.log.Println("[ERROR] validating order", err)
			http.Error(
				w, fmt.Sprintf(`{"error": "Error validating order: %s"}`, err),
				http.StatusBadRequest,
			)
			return
		}

		// add validated data
		ctx := context.WithValue(r.Context(), OrderCTXKey, order)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (h *OrderHandler) MiddlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authString := r.Header.Get("Authorization")
		if authString == "" {
			ErrUnauthorized(w, utils.ErrMissingAuthToken)
			return
		}

		tokenString, err := utils.ExtractToken(authString)
		if err != nil {
			ErrUnauthorized(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), AuthCTXKey, tokenString)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ErrUnauthorized is a helper for consistent unauthorized responses
func ErrUnauthorized(w http.ResponseWriter, err error) {
	errRes := fmt.Sprintf(`{"error": "%v"}`, err.Error())
	http.Error(w, errRes, http.StatusUnauthorized)
}
