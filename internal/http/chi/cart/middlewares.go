package cart

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rovilay/ecommerce-service/domains/cart/models"
	"github.com/rovilay/ecommerce-service/utils"
)

type contextKey string

const CartCTXKey contextKey = "cart_item_payload"
const AuthCTXKey contextKey = "auth_token"

func (h *CartHandler) MiddlewareValidateCartItem(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		item := &models.CartItem{}

		if err := item.FromJSON(r.Body); err != nil {
			h.log.Println("[ERROR] deserializing cart item", err)
			http.Error(w, `{"error": "failed to read payload"}`, http.StatusBadRequest)
			return
		}

		// validate url
		// err = inv.Validate()
		// if err != nil {
		// 	h.log.Println("[ERROR] validating inventory", err)
		// 	http.Error(
		// 		w, fmt.Sprintf(`{"error": "Error validating inventory: %s"}`, err),
		// 		http.StatusBadRequest,
		// 	)
		// 	return
		// }

		// add validated data
		ctx := context.WithValue(r.Context(), CartCTXKey, item)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (h *CartHandler) MiddlewareAuth(next http.Handler) http.Handler {
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
