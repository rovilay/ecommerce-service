package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/rovilay/ecommerce-service/config"
	"github.com/rovilay/ecommerce-service/domains/auth"
	externalservices "github.com/rovilay/ecommerce-service/domains/order/external-services"
	"github.com/rovilay/ecommerce-service/domains/order/repository"
	"github.com/rovilay/ecommerce-service/domains/order/service"
	httpOrder "github.com/rovilay/ecommerce-service/internal/http/chi/order"
	"github.com/rs/zerolog"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout).With().Str("component", "order-service:main").Timestamp().Logger()

	// notify context of os.Interrupt signal
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// add context to logger
	ctx = logger.WithContext(ctx)

	// load env
	// envPath, err := filepath.Abs("./.env")
	// if err != nil {
	// 	logger.Fatal().Err(err).Msg("Error resolving .env path")
	// }

	// err = godotenv.Load(envPath)
	// if err != nil {
	//	// logger.Fatal().Err(err).Msg("Error loading .env file")
	// 	logger.Err(err).Msg("error loading .env file")
	// }

	// load the config
	c := config.LoadOrderConfig(&logger)

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

	// connect to redis
	cache := redis.NewClient(&redis.Options{
		Addr: c.RedisURL,
	})
	err = cache.Ping(ctx).Err()
	if err != nil {
		logger.Fatal().Err(err).Msg(fmt.Sprintf("failed to connect to redis: %s", c.RedisURL))
	}
	defer func() {
		if err := cache.Close(); err != nil {
			logger.Err(err).Msg("failed to close redis")
		}
	}()

	repo := repository.NewPostgresOrderRepository(ctx, db, &logger)
	authService := auth.NewAuthService(cache, c.AuthSecret, time.Hour*10)
	inventoryService := externalservices.NewHTTPInventoryService(c.InventoryHttpBaseURL)
	prdService := externalservices.NewHTTPProductService(c.ProdHttpBaseURL)
	cartService := externalservices.NewHTTPCartService(c.CartHttpBaseURL)
	service := service.NewOrderService(repo, authService, inventoryService, prdService, cartService, &logger)
	app := httpOrder.NewOrderApp(service, &c, &logger)
	if err = app.Start(ctx); err != nil {
		logger.Fatal().Err(err).Msg("failed to start app")
	}
}
