package inventory

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rovilay/ecommerce-service/config"
	"github.com/rovilay/ecommerce-service/domains/inventory/service"
	"github.com/rs/zerolog"
)

type InventoryApp struct {
	router  http.Handler
	config  *config.InventoryConfig
	log     *zerolog.Logger
	service *service.InventoryService
}

func NewInventoryApp(s *service.InventoryService, c *config.InventoryConfig, log *zerolog.Logger) *InventoryApp {
	logger := log.With().Str("package", "InventoryApp").Logger()

	app := &InventoryApp{
		log:     &logger,
		config:  c,
		service: s,
	}

	app.loadRoutes()

	return app
}

func (a *InventoryApp) Start(ctx context.Context) error {
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
