package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/rovilay/ecommerce-service/internal/http/chi/product"
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

	app := product.NewProductApp(product.LoadConfig(), &logger)

	if err = app.Start(ctx); err != nil {
		logger.Fatal().Err(err).Msg("failed to start app")
	}
}
