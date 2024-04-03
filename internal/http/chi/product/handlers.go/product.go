package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rovilay/ecommerce-service/domains/product"
)

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.log.Println("bad ID param: ", err)
		http.Error(w, `{"error": "failed to convert product ID param"}`, http.StatusBadRequest)
		return
	}

	prd, err := h.service.GetProduct(r.Context(), productID)
	if errors.Is(err, product.ErrNotExist) {
		http.Error(w, `{"error": "product not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		h.log.Println("failed to get product: ", err)
		http.Error(w, `{"error": "failed to get product"}`, http.StatusInternalServerError)
		return
	}

	if err := prd.ToJSON(w); err != nil {
		h.log.Println("failed to marshal: ", err)
		http.Error(w, `{"error": "failed to marshal"}`, http.StatusInternalServerError)
		return
	}
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		h.log.Err(err)
		limit = 50
	}
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		h.log.Err(err)
		offset = 0
	}

	res, err := h.service.ListProducts(r.Context(), limit, offset)
	if err != nil {
		h.log.Println("failed to list products: ", err)
		http.Error(w, `{"error": "failed to list products"}`, http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.log.Println("failed to marshal: ", err)
		http.Error(w, `{"error": "failed to marshal"}`, http.StatusInternalServerError)
		return
	}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	data := r.Context().Value(PrdCTXKey).(*product.Product)

	newPrd, err := h.service.CreateProduct(r.Context(), data)
	if err != nil {
		h.log.Println("failed to create product: ", err)
		http.Error(w, `{"error": "failed to create product"}`, http.StatusInternalServerError)
		return
	}

	if err := newPrd.ToJSON(w); err != nil {
		h.log.Println("failed to marshal: ", err)
		http.Error(w, `{"error": "failed to marshal"}`, http.StatusInternalServerError)
		return
	}
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.log.Println("bad ID param: ", err)
		http.Error(w, `{"error": "failed to convert product ID param"}`, http.StatusBadRequest)
		return
	}

	data := r.Context().Value(PrdCTXKey).(*product.Product)

	prd, err := h.service.UpdateProduct(r.Context(), productID, data)
	if errors.Is(err, product.ErrNotExist) {
		http.Error(w, `{"error": "product not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		h.log.Println("failed to update product: ", err)
		http.Error(w, `{"error": "failed to update product"}`, http.StatusInternalServerError)
		return
	}

	if err := prd.ToJSON(w); err != nil {
		h.log.Println("failed to marshal: ", err)
		http.Error(w, `{"error": "failed to marshal"}`, http.StatusInternalServerError)
		return
	}
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.log.Println("bad ID param: ", err)
		http.Error(w, `{"error": "failed to convert product ID param"}`, http.StatusBadRequest)
		return
	}

	err = h.service.DeleteProduct(r.Context(), productID)
	if errors.Is(err, product.ErrNotExist) {
		http.Error(w, `{"error": "product not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		h.log.Println("failed to delete by id: ", err)
		http.Error(w, `{"error": "failed to delete resource"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) SearchProducts(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("q")
	if searchTerm == "" {
		http.Error(w, `{"error": "search term empty"}`, http.StatusBadGateway)
		return
	}
	res, err := h.service.SearchProductsByName(r.Context(), searchTerm)
	if err != nil {
		h.log.Println("product search failed: ", err)
		http.Error(w, `{"error": "product search failed"}`, http.StatusInternalServerError)
		return
	}

	var response struct {
		Result []*product.Product `json:"result"`
	}

	response.Result = res

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Println("failed to marshal: ", err)
		http.Error(w, `{"error": "failed to marshal"}`, http.StatusInternalServerError)
		return
	}
}
