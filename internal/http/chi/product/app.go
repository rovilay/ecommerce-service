package product

import (
	"context"
	"fmt"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type ProductApp struct {
	router http.Handler
	config Config
	db     *sqlx.DB
	log    *zerolog.Logger
}

func NewProductApp(c Config, log *zerolog.Logger) *ProductApp {
	appLogger := log.With().Str("package", "productApp").Logger()

	connectionURL := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/ecommerce-service-products?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort,
	)

	db, err := sqlx.Connect("pgx", connectionURL)
	if err != nil {
		appLogger.Fatal().Err(err).Msg(fmt.Sprintf("failed to connect to DB %s", connectionURL))
	}

	app := &ProductApp{
		config: c,
		log:    log,
		db:     db,
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
