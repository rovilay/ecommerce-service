package product

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	handler "github.com/rovilay/ecommerce-service/internal/http/chi/product/handlers.go"
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

	router.Route("/products", a.loadProductRoutes)

	// CORS configuration
	corsRouter := cors.Default().Handler(router)

	a.router = corsRouter
}

func (a *ProductApp) loadProductRoutes(router chi.Router) {
	prdHandler := handler.NewProductHandler(a.service, a.log)

	router.Get("/", prdHandler.ListProducts)
	router.Get("/{id}", prdHandler.GetProduct)

	router.Group(func(r chi.Router) {
		r.Use(prdHandler.MiddlewareValidateProduct)
		r.Post("/", prdHandler.CreateProduct)
		r.Put("/{id}", prdHandler.UpdateProduct)
	})

	router.Delete("/{id}", prdHandler.DeleteProduct)
	router.Get("/search", prdHandler.SearchProducts)

	router.Get("/categories", prdHandler.ListCategories)
	router.Get("/categories/{id}", prdHandler.GetCategory)

	router.Group(func(r chi.Router) {
		r.Use(prdHandler.MiddlewareValidateCategory)
		r.Post("/categories", prdHandler.CreateCategory)
		r.Put("/categories/{id}", prdHandler.UpdateCategory)
	})

	router.Get("/categories/search", prdHandler.SearchCategories)
}
