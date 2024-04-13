package cart

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

func (a *CartApp) loadRoutes() {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var res struct {
			Message string `json:"message"`
		}

		res.Message = "Welcome to cart service"

		msg, err := json.Marshal(res)
		if err != nil {
			fmt.Println("failed to marshall ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(msg)
	})

	router.Route("/cart", a.loadCartRoutes)

	// CORS configuration
	corsRouter := cors.Default().Handler(router)

	a.router = corsRouter
}

func (a *CartApp) loadCartRoutes(router chi.Router) {
	h := NewCartHandler(a.service, a.log)

	router.Group(func(r chi.Router) {
		r.Use(h.MiddlewareAuth)
		r.Get("/", h.GetCart)
		r.Delete("/items/{id}", h.RemoveItem)
		r.Delete("/", h.ClearCart)
	})

	router.Group(func(r chi.Router) {
		r.Use(h.MiddlewareAuth)
		r.Use(h.MiddlewareValidateCartItem)
		r.Post("/", h.AddItem)
		r.Put("/items/{id}", h.UpdateCartItemQuantity)
	})
}
