package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/rovilay/ecommerce-service/config"
	"github.com/rovilay/ecommerce-service/domains/inventory/repository"
	"github.com/rovilay/ecommerce-service/domains/inventory/service"
	inventoryHttp "github.com/rovilay/ecommerce-service/internal/http/chi/inventory"
	"github.com/rs/zerolog"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout).With().Str("component", "inventory-service:main").Timestamp().Logger()

	// notify context of os.Interrupt signal
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// add context to logger
	ctx = logger.WithContext(ctx)

	// load env
	envPath, err := filepath.Abs("./.env")
	if err != nil {
		logger.Fatal().Err(err).Msg("Error resolving .env path")
	}

	err = godotenv.Load(envPath)
	if err != nil {
		// logger.Fatal().Err(err).Msg("Error loading .env file")
		logger.Err(err).Msg("error loading .env file")
	}

	// load the config
	c := config.LoadInventoryConfig()

	// connect to DB
	db, err := sqlx.Connect("pgx", c.DBURL)
	if err != nil {
		logger.Fatal().Err(err).Msg(fmt.Sprintf("failed to connect to DB %s", c.DBURL))
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Err(err).Msg("failed to close postgres")
		}
	}()

	repo := repository.NewPostgresInventoryRepository(ctx, db, &logger)
	service := service.NewInventoryService(repo, &logger)
	app := inventoryHttp.NewInventoryApp(service, &c, &logger)

	if err = app.Start(ctx); err != nil {
		logger.Fatal().Err(err).Msg("failed to start app")
	}
}
