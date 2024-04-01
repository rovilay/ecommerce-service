package product

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

func (a *ProductApp) loadRoutes() {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var res struct {
			Message string
		}

		res.Message = "Welcome to product service"

		msg, err := json.Marshal(res)
		if err != nil {
			fmt.Println("failed to marshall ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(msg)
	})

	// router.Route("/products", a.loadProductRoutes)

	// CORS configuration
	corsRouter := cors.Default().Handler(router)

	a.router = corsRouter
}
