package order

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rovilay/ecommerce-service/domains/order"
	"github.com/rovilay/ecommerce-service/domains/order/models"
	"github.com/rovilay/ecommerce-service/domains/order/service"
	"github.com/rs/zerolog"
)

type OrderHandler struct {
	service *service.OrderService
	log     *zerolog.Logger
}

type successOperation struct {
	Success string `json:"success"`
}

var defaultSuccessRes = successOperation{Success: "operation successful!"}

func NewOrderHandler(s *service.OrderService, l *zerolog.Logger) *OrderHandler {
	logger := l.With().Str("component", "OrderHandler").Logger()

	return &OrderHandler{
		service: s,
		log:     &logger,
	}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("method", "CreateOrder").Logger()
	authToken := r.Context().Value(AuthCTXKey).(string)
	data := r.Context().Value(OrderCTXKey).(*models.Order)
	fromCart := len(data.OrderItems) == 0

	order, err := h.service.CreateOrder(r.Context(), authToken, data, fromCart)
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(&order); err != nil {
		h.sendError(w, err, "failed to marshal", 0, &log)
		return
	}
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("method", "GetOrder").Logger()
	authToken := r.Context().Value(AuthCTXKey).(string)

	orderID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, err, "failed to convert product ID param", http.StatusBadRequest, &log)
		return
	}

	order, err := h.service.GetOrder(r.Context(), authToken, orderID)
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	if err = json.NewEncoder(w).Encode(&order); err != nil {
		h.sendError(w, err, "failed to marshal", 0, &log)
		return
	}
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("method", "GetOrders").Logger()
	authToken := r.Context().Value(AuthCTXKey).(string)

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		log.Err(err)
		limit = 50
	}
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		log.Err(err)
		offset = 0
	}

	orders, err := h.service.GetUserOrders(r.Context(), authToken, limit, offset)
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	if err = json.NewEncoder(w).Encode(&orders); err != nil {
		h.sendError(w, err, "failed to marshal", 0, &log)
		return
	}
}

func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("method", "UpdateOrderStatus").Logger()
	authToken := r.Context().Value(AuthCTXKey).(string)

	orderID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, err, "failed to convert product ID param", http.StatusBadRequest, &log)
		return
	}

	var payload struct {
		Status string `json:"status"`
	}

	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.log.Println("[ERROR] deserializing order status", err)
		http.Error(w, `{"error": "failed to read payload"}`, http.StatusBadRequest)
		return
	}

	status := models.OrderStatus(payload.Status)

	err = h.service.UpdateOrderStatus(r.Context(), authToken, orderID, status)
	if err != nil {
		h.sendError(w, err, "", http.StatusBadRequest, &log)
		return
	}

	if err = json.NewEncoder(w).Encode(defaultSuccessRes); err != nil {
		h.sendError(w, err, "failed to marshal", 0, &log)
		return
	}
}

func (h *OrderHandler) sendError(w http.ResponseWriter, err error, errMsg string, statusCode int, log *zerolog.Logger) {
	log.Err(err)
	if errMsg == "" {
		errMsg = err.Error()
	}

	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	errRes := fmt.Sprintf(`{"error": "%v"}`, errMsg)

	if errors.Is(err, order.ErrInvalidProduct) || errors.Is(err, order.ErrInsufficientStock) ||
		errors.Is(err, order.ErrInvalidQuantity) || errors.Is(err, order.ErrDuplicateEntry) ||
		errors.Is(err, order.ErrForeignKeyViolation) {
		http.Error(w, errRes, http.StatusBadRequest)
		return
	} else if errors.Is(err, order.ErrInvalidJWToken) {
		http.Error(w, errRes, http.StatusUnauthorized)
		return
	} else if errors.Is(err, order.ErrNotFound) || errors.Is(err, order.ErrItemNotFound) {
		http.Error(w, errRes, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, errRes, statusCode)
		return
	}
}
