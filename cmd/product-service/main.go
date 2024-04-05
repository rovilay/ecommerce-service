package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/rovilay/ecommerce-service/config"
	"github.com/rovilay/ecommerce-service/domains/product"
	productHttp "github.com/rovilay/ecommerce-service/internal/http/chi/product"
	"github.com/rs/zerolog"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stderr).With().Str("component", "main").Timestamp().Logger()

	// notify context of os.Interrupt signal
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// attach logger to context
	ctx = logger.WithContext(ctx)

	envPath, err := filepath.Abs("./config/.env")
	if err != nil {
		logger.Fatal().Err(err).Msg("Error loading .env file")
	}

	// Load .env file from the current directory
	err = godotenv.Load(envPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("Error loading .env file")
	}

	// load config
	c := config.LoadProductConfig()

	db, err := sqlx.Connect("pgx", c.DBURL)
	if err != nil {
		logger.Fatal().Err(err).Msg(fmt.Sprintf("failed to connect to DB %s", c.DBURL))
	}

	postgresRepo := product.NewPostgresRepository(db, &logger)
	productService := product.NewService(&postgresRepo)
	app := productHttp.NewProductApp(db, productService, &c, &logger)

	if err = app.Start(ctx); err != nil {
		logger.Fatal().Err(err).Msg("failed to start app")
	}
}
