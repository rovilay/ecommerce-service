package inventory

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

func (a *InventoryApp) loadRoutes() {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var res struct {
			Message string `json:"message"`
		}

		res.Message = "Welcome to inventory service"

		msg, err := json.Marshal(res)
		if err != nil {
			fmt.Println("failed to marshall ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(msg)
	})

	router.Route("/api/v1/inventory", a.loadInventoryRoutes)

	// CORS configuration
	corsRouter := cors.Default().Handler(router)

	a.router = corsRouter
}

func (a *InventoryApp) loadInventoryRoutes(router chi.Router) {
	h := NewInventoryHandler(a.service, a.log)

	router.Get("/products/{id}", h.GetInventory)
	router.Get("/products/{id}/available", h.CheckAvailability)

	router.Put("/products/{id}/increase", h.IncrementInventory)
	router.Put("/products/{id}/decrease", h.DecrementInventory)
}
