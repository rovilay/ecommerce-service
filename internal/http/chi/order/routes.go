package order

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

func (a *OrderApp) loadRoutes() {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var res struct {
			Message string `json:"message"`
		}

		res.Message = "Welcome to order service"

		msg, err := json.Marshal(res)
		if err != nil {
			fmt.Println("failed to marshall ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(msg)
	})

	router.Route("/orders", a.loadOrderRoutes)

	// CORS configuration
	corsRouter := cors.Default().Handler(router)

	a.router = corsRouter
}

func (a *OrderApp) loadOrderRoutes(router chi.Router) {
	h := NewOrderHandler(a.service, a.log)

	router.Group(func(r chi.Router) {
		r.Use(h.MiddlewareAuth)
		r.Get("/", h.GetOrders)
		r.Get("/{id}", h.GetOrder)
		r.Put("/{id}/status", h.UpdateOrderStatus)
	})

	router.Group(func(r chi.Router) {
		r.Use(h.MiddlewareAuth)
		r.Use(h.MiddlewareValidateOrderItems)
		r.Post("/", h.CreateOrder)
	})
}
