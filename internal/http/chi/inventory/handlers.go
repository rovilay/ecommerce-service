package inventory

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rovilay/ecommerce-service/domains/inventory"
	"github.com/rovilay/ecommerce-service/domains/inventory/service"
	"github.com/rs/zerolog"
)

type InventoryHandler struct {
	service *service.InventoryService
	log     *zerolog.Logger
}

func NewInventoryHandler(s *service.InventoryService, l *zerolog.Logger) *InventoryHandler {
	logger := l.With().Str("component", "InventoryHandler").Logger()

	return &InventoryHandler{
		service: s,
		log:     &logger,
	}
}

type successOperation struct {
	Success string `json:"success"`
}

var defaultSuccessRes = successOperation{Success: "operation successful!"}

func (h *InventoryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("method", "GetInventory").Logger()

	productID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, err, "failed to convert product ID param", http.StatusBadRequest, &log)
		return
	}

	ivntry, err := h.service.GetInventoryByProductID(r.Context(), productID)
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	if err := ivntry.ToJSON(w); err != nil {
		h.sendError(w, err, "failed to marshal", 0, &log)
		return
	}
}

func (h *InventoryHandler) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("method", "CheckAvailability").Logger()

	productID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, err, "failed to convert product ID param", http.StatusBadRequest, &log)
		return
	}

	var availableRes struct {
		Available bool `json:"available"`
	}

	availableRes.Available = false

	qty := r.URL.Query().Get("qty")
	if qty == "" {
		h.sendError(w, err, "qty query param missing", 0, &log)
		return
	}
	quantity, err := strconv.Atoi(qty)
	if err != nil {
		h.sendError(w, err, "failed to convert qty value", http.StatusBadRequest, &log)
		return
	}

	available, err := h.service.CheckAvailability(r.Context(), productID, uint(quantity))
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	availableRes.Available = available

	if err = json.NewEncoder(w).Encode(availableRes); err != nil {
		h.sendError(w, err, "failed to marshal", 0, &log)
		return
	}
}

func (h *InventoryHandler) DecrementInventory(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("method", "DecrementInventory").Logger()

	productID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, err, "failed to convert product ID param", http.StatusBadRequest, &log)
		return
	}

	var payload struct {
		Quantity int `json:"quantity"`
	}

	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.sendError(w, err, "failed to decode payload", http.StatusBadRequest, &log)
		return
	}

	err = h.service.DecrementInventory(r.Context(), productID, uint(payload.Quantity))
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	if err = json.NewEncoder(w).Encode(defaultSuccessRes); err != nil {
		h.sendError(w, err, "failed to marshal", 0, &log)
		return
	}
}

func (h *InventoryHandler) IncrementInventory(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("method", "IncrementInventory").Logger()

	productID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, err, "failed to convert product ID param", http.StatusBadRequest, &log)
		return
	}

	var payload struct {
		Quantity int `json:"quantity"`
	}

	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.sendError(w, err, "failed to decode payload", http.StatusBadRequest, &log)
		return
	}

	err = h.service.IncrementInventory(r.Context(), productID, uint(payload.Quantity))
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	if err = json.NewEncoder(w).Encode(defaultSuccessRes); err != nil {
		h.sendError(w, err, "failed to marshal", 0, &log)
		return
	}
}

func (h *InventoryHandler) sendError(w http.ResponseWriter, err error, errMsg string, statusCode int, log *zerolog.Logger) {
	log.Err(err)

	if errMsg == "" {
		errMsg = err.Error()
	}

	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	errRes := fmt.Sprintf(`{"error": "%v"}`, errMsg)

	if errors.Is(err, inventory.ErrInvalidProduct) || errors.Is(err, inventory.ErrInsufficientStock) ||
		errors.Is(err, inventory.ErrInvalidQuantity) || errors.Is(err, inventory.ErrDuplicateEntry) ||
		errors.Is(err, inventory.ErrForeignKeyViolation) {
		http.Error(w, errRes, http.StatusBadRequest)
		return
	} else if errors.Is(err, inventory.ErrNotFound) {
		http.Error(w, errRes, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, errRes, statusCode)
		return
	}
}
