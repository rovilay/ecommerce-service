package inventory

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rovilay/ecommerce-service/domains/inventory"
	"github.com/rovilay/ecommerce-service/domains/inventory/model"
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

func (h *InventoryHandler) CreateInventory(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("method", "CreateInventory").Logger()
	data := r.Context().Value(InvCTXKey).(*model.InventoryItem)

	newInv, err := h.service.CreateInventoryItem(r.Context(), data.ProductID, data.Quantity)
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	if err := newInv.ToJSON(w); err != nil {
		h.sendError(w, err, "failed to marshal", 0, &log)
		return
	}
}

func (h *InventoryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("method", "GetInventory").Logger()

	productID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, err, "failed to convert product ID param", http.StatusBadRequest, &log)
		return
	}

	prd, err := h.service.GetInventoryByProductID(r.Context(), productID)
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	if err := prd.ToJSON(w); err != nil {
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

	if errors.Is(err, inventory.ErrInvalidProduct) || errors.Is(err, inventory.ErrInvalidQuantity) {
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
