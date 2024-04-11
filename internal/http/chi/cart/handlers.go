package cart

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rovilay/ecommerce-service/domains/cart"
	"github.com/rovilay/ecommerce-service/domains/cart/models"
	"github.com/rovilay/ecommerce-service/domains/cart/service"
	"github.com/rs/zerolog"
)

type CartHandler struct {
	service *service.CartService
	log     *zerolog.Logger
}

func NewCartHandler(s *service.CartService, l *zerolog.Logger) *CartHandler {
	logger := l.With().Str("component", "CartHandler").Logger()

	return &CartHandler{
		service: s,
		log:     &logger,
	}
}

type successOperation struct {
	Success string `json:"success"`
}

var defaultSuccessRes = successOperation{Success: "operation successful!"}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("method", "GetCart").Logger()
	authToken := r.Context().Value(AuthCTXKey).(string)

	cart, err := h.service.GetCart(r.Context(), authToken)
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	if err := cart.ToJSON(w); err != nil {
		h.sendError(w, err, "failed to marshal", 0, &log)
		return
	}
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("method", "AddItem").Logger()
	authToken := r.Context().Value(AuthCTXKey).(string)
	data := r.Context().Value(CartCTXKey).(*models.CartItem)

	if data.Quantity <= 0 {
		h.sendError(w, cart.ErrInvalidQuantity, "", 0, &log)
		return
	}

	newItem, err := h.service.AddItemToCart(r.Context(), authToken, *data)
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	w.WriteHeader(http.StatusCreated)

	if err := newItem.ToJSON(w); err != nil {
		h.sendError(w, err, "failed to marshal", 0, &log)
		return
	}
}

func (h *CartHandler) UpdateCartItemQuantity(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("method", "UpdateCartItemQuantity").Logger()
	authToken := r.Context().Value(AuthCTXKey).(string)

	cartItemID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, err, "failed to convert product ID param", http.StatusBadRequest, &log)
		return
	}

	data := r.Context().Value(CartCTXKey).(*models.CartItem)

	if data.Quantity < 0 {
		err = cart.ErrInvalidQuantity
	} else if data.Quantity == 0 {
		err = h.service.RemoveItemFromCart(r.Context(), authToken, cartItemID)
	} else {
		data.ID = cartItemID
		err = h.service.UpdateCartItemQuantity(r.Context(), authToken, *data)
	}

	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	if err = json.NewEncoder(w).Encode(defaultSuccessRes); err != nil {
		h.sendError(w, err, "failed to marshal", 0, &log)
		return
	}
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("method", "RemoveItem").Logger()
	authToken := r.Context().Value(AuthCTXKey).(string)
	cartItemID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, err, "failed to convert product ID param", http.StatusBadRequest, &log)
		return
	}

	err = h.service.RemoveItemFromCart(r.Context(), authToken, cartItemID)
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("method", "ClearCart").Logger()
	authToken := r.Context().Value(AuthCTXKey).(string)

	err := h.service.ClearCart(r.Context(), authToken)
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CartHandler) sendError(w http.ResponseWriter, err error, errMsg string, statusCode int, log *zerolog.Logger) {
	log.Err(err)
	if errMsg == "" {
		errMsg = err.Error()
	}

	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	errRes := fmt.Sprintf(`{"error": "%v"}`, errMsg)

	if errors.Is(err, cart.ErrInvalidProduct) || errors.Is(err, cart.ErrInsufficientStock) ||
		errors.Is(err, cart.ErrInvalidQuantity) || errors.Is(err, cart.ErrDuplicateEntry) ||
		errors.Is(err, cart.ErrForeignKeyViolation) {
		http.Error(w, errRes, http.StatusBadRequest)
		return
	} else if errors.Is(err, cart.ErrInvalidJWToken) {
		http.Error(w, errRes, http.StatusUnauthorized)
		return
	} else if errors.Is(err, cart.ErrNotFound) {
		http.Error(w, errRes, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, errRes, statusCode)
		return
	}
}
