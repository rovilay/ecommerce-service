package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rovilay/ecommerce-service/domains/product"
)

func (h *ProductHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	categoryID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.log.Println("bad ID param: ", err)
		http.Error(w, `{"error": "failed to convert product ID param"}`, http.StatusBadRequest)
		return
	}

	prd, err := h.service.GetCategory(r.Context(), categoryID)
	if errors.Is(err, product.ErrNotExist) {
		http.Error(w, `{"error": "category resource not found"}`, http.StatusNotFound)
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

func (h *ProductHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
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

	res, err := h.service.ListCategories(r.Context(), limit, offset)
	if err != nil {
		h.log.Println("failed to list categories: ", err)
		http.Error(w, `{"error": "failed to list categories"}`, http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.log.Println("failed to marshal: ", err)
		http.Error(w, `{"error": "failed to marshal"}`, http.StatusInternalServerError)
		return
	}
}

func (h *ProductHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	data := r.Context().Value(CategoryCTXKey).(*product.Category)

	newPrd, err := h.service.CreateCategory(r.Context(), data)
	if err != nil {
		h.log.Println("failed to create category: ", err)
		http.Error(w, `{"error": "failed to create category"}`, http.StatusInternalServerError)
		return
	}

	if err := newPrd.ToJSON(w); err != nil {
		h.log.Println("failed to marshal: ", err)
		http.Error(w, `{"error": "failed to marshal"}`, http.StatusInternalServerError)
		return
	}
}

func (h *ProductHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	ctgryID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.log.Println("bad ID param: ", err)
		http.Error(w, `{"error": "failed to convert category ID param"}`, http.StatusBadRequest)
		return
	}

	data := r.Context().Value(CategoryCTXKey).(*product.Category)

	prd, err := h.service.UpdateCategory(r.Context(), ctgryID, data)
	if errors.Is(err, product.ErrNotExist) {
		http.Error(w, `{"error": "category resource not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		h.log.Println("failed to update category: ", err)
		http.Error(w, `{"error": "failed to update category"}`, http.StatusInternalServerError)
		return
	}

	if err := prd.ToJSON(w); err != nil {
		h.log.Println("failed to marshal: ", err)
		http.Error(w, `{"error": "failed to marshal"}`, http.StatusInternalServerError)
		return
	}
}

func (h *ProductHandler) SearchCategories(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("q")
	if searchTerm == "" {
		http.Error(w, `{"error": "search term empty"}`, http.StatusBadGateway)
		return
	}
	res, err := h.service.SearchCategoriesByName(r.Context(), searchTerm)
	if err != nil {
		h.log.Println("category search failed: ", err)
		http.Error(w, `{"error": "category search failed"}`, http.StatusInternalServerError)
		return
	}

	var response struct {
		Result []*product.Category `json:"result"`
	}

	response.Result = res

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Println("failed to marshal: ", err)
		http.Error(w, `{"error": "failed to marshal"}`, http.StatusInternalServerError)
		return
	}
}
