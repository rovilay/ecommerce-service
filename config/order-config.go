package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/rs/zerolog"
)

type OrderConfig struct {
	ServerPort           uint16
	DBURL                string
	InventoryHttpBaseURL string
	ProdHttpBaseURL      string
	CartHttpBaseURL      string
	AuthSecret           string
	RedisURL             string
}

func LoadOrderConfig(log *zerolog.Logger) OrderConfig {
	cfg := OrderConfig{
		ServerPort: 3000,
	}

	if serverPort, exists := os.LookupEnv("ORDER_SERVER_PORT"); exists {
		if port, err := strconv.ParseUint(serverPort, 10, 16); err == nil {
			cfg.ServerPort = uint16(port)
		}
	}

	if url, exists := os.LookupEnv("REDIS_URL"); exists {
		cfg.RedisURL = url
	}

	if url, exists := os.LookupEnv("DB_URL"); exists {
		cfg.DBURL = url
	}

	if secret, exists := os.LookupEnv("USER_AUTH_SECRET"); exists {
		cfg.AuthSecret = secret
	} else {
		log.Fatal().Err(errors.New("USER_AUTH_SECRET is required")).Msg("failed to load config")
	}

	if url, exists := os.LookupEnv("PRODUCT_BASE_URL"); exists {
		cfg.ProdHttpBaseURL = url
	} else {
		log.Fatal().Err(errors.New("PRODUCT_BASE_URL is required")).Msg("failed to load config")
	}

	if url, exists := os.LookupEnv("INVENTORY_BASE_URL"); exists {
		cfg.InventoryHttpBaseURL = url
	} else {
		log.Fatal().Err(errors.New("INVENTORY_BASE_URL is required")).Msg("failed to load config")
	}

	if url, exists := os.LookupEnv("CART_BASE_URL"); exists {
		cfg.CartHttpBaseURL = url
	} else {
		log.Fatal().Err(errors.New("CART_BASE_URL is required")).Msg("failed to load config")
	}

	return cfg
}
