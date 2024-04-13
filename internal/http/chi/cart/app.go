package cart

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rovilay/ecommerce-service/config"
	"github.com/rovilay/ecommerce-service/domains/cart/service"
	"github.com/rs/zerolog"
)

type CartApp struct {
	router  http.Handler
	config  *config.CartConfig
	log     *zerolog.Logger
	service *service.CartService
}

func NewCartApp(s *service.CartService, c *config.CartConfig, log *zerolog.Logger) *CartApp {
	logger := log.With().Str("package:cart", "CartApp").Logger()

	app := &CartApp{
		log:     &logger,
		config:  c,
		service: s,
	}

	app.loadRoutes()

	return app
}

func (a *CartApp) Start(ctx context.Context) error {
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
