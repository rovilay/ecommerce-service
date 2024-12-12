package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rovilay/ecommerce-service/common/events"
	"github.com/rovilay/ecommerce-service/config"
	"github.com/rovilay/ecommerce-service/domains/product"
	productHttp "github.com/rovilay/ecommerce-service/internal/http/chi/product"
	"github.com/rs/zerolog"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stderr).With().Str("component", "product-service:main").Timestamp().Logger()

	// notify context of os.Interrupt signal
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// attach logger to context
	ctx = logger.WithContext(ctx)

	// envPath, err := filepath.Abs("./.env")
	// if err != nil {
	// 	logger.Fatal().Err(err).Msg("Error resolving .env path")
	// }

	// // Load .env file from the current directory
	// err = godotenv.Load(envPath)
	// if err != nil {
	// 	// logger.Fatal().Err(err).Msg("Error loading .env file")
	// 	logger.Err(err).Msg("error loading .env file")
	// }

	// load config
	c := config.LoadProductConfig()

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

	// connect to rabbitmq
	// conn, err := events.ConnectRabbit(c.RABBITMQ_USER, c.RABBITMQ_PASSWORD, c.RABBITMQ_HOST, c.RABBITMQ_PORT)
	conn, err := events.ConnectRabbit(c.RABBITMQ_URL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to rabbitMq")
	}

	rabbitClient, err := events.NewRabbitClient(conn, events.Product)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create rabbit client")
	}

	logger.Info().Msg("Connected to rabbit client")

	defer rabbitClient.Close()

	postgresRepo := product.NewPostgresRepository(ctx, db, logger)
	productService, err := product.NewService(postgresRepo, rabbitClient, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("product.NewService: something went wrong")
	}

	app := productHttp.NewProductApp(productService, &c, &logger)

	if err = app.Start(ctx); err != nil {
		logger.Fatal().Err(err).Msg("failed to start app")
	}
}
