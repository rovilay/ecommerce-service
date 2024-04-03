package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rovilay/ecommerce-service/domains/product"
)

type contextKey string

const PrdCTXKey contextKey = "product"
const CategoryCTXKey contextKey = "category"

func (h *ProductHandler) MiddlewareValidateProduct(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prd := &product.Product{}

		err := prd.FromJSON(r.Body)
		if err != nil {
			h.log.Println("[ERROR] deserializing product", err)
			http.Error(w, `{"error": "failed to read product"}`, http.StatusBadRequest)
			return
		}

		// validate url
		err = prd.Validate()
		if err != nil {
			h.log.Println("[ERROR] validating product", err)
			http.Error(
				w, fmt.Sprintf(`{"error": "Error valdating product: %s"}`, err),
				http.StatusBadRequest,
			)
			return
		}

		// add validated data
		ctx := context.WithValue(r.Context(), PrdCTXKey, prd)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (h *ProductHandler) MiddlewareValidateCategory(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := &product.Category{}

		err := c.FromJSON(r.Body)
		if err != nil {
			h.log.Println("[ERROR] deserializing category", err)
			http.Error(w, `{"error": "failed to read category"}`, http.StatusBadRequest)
			return
		}

		// validate url
		err = c.Validate()
		if err != nil {
			h.log.Println("[ERROR] validating category", err)
			http.Error(
				w, fmt.Sprintf(`{"error": "Error valdating category: %s"}`, err),
				http.StatusBadRequest,
			)
			return
		}

		// add validated data
		ctx := context.WithValue(r.Context(), CategoryCTXKey, c)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
