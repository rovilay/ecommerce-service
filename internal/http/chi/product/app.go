package product

import (
	"context"
	"fmt"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rovilay/ecommerce-service/config"
	"github.com/rovilay/ecommerce-service/domains/product"
	"github.com/rs/zerolog"
)

type ProductApp struct {
	router  http.Handler
	config  *config.ProductConfig
	db      *sqlx.DB
	log     *zerolog.Logger
	service *product.Service
}

func NewProductApp(db *sqlx.DB, s *product.Service, c *config.ProductConfig, log *zerolog.Logger) *ProductApp {
	appLogger := log.With().Str("package", "productApp").Logger()

	app := &ProductApp{
		config:  c,
		log:     &appLogger,
		db:      db,
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

	// ping db
	if err := a.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}

	defer func() {
		if err := a.db.Close(); err != nil {
			a.log.Err(err).Msg("failed to close postgres")
		}
	}()

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
