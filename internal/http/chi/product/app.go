package product

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rovilay/ecommerce-service/config"
	"github.com/rovilay/ecommerce-service/domains/product"
	"github.com/rs/zerolog"
)

type ProductApp struct {
	router  http.Handler
	config  *config.ProductConfig
	log     *zerolog.Logger
	service *product.Service
}

func NewProductApp(s *product.Service, c *config.ProductConfig, log *zerolog.Logger) *ProductApp {
	appLogger := log.With().Str("package", "productApp").Logger()

	app := &ProductApp{
		config:  c,
		log:     &appLogger,
		service: s,
	}

	app.loadRoutes()

	return app
}

func (a *ProductApp) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.ServerPort),
		Handler: a.router,
	}

	a.log.Println("Starting Server on port: ", server.Addr)

	ch := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		return server.Shutdown(timeout)
	}
}
