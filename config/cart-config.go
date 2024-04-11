package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/rs/zerolog"
)

type CartConfig struct {
	ServerPort uint16
	DBURL      string
	AuthSecret string
	RedisURL   string
}

func LoadCartConfig(log *zerolog.Logger) CartConfig {
	cfg := CartConfig{
		ServerPort: 3000,
	}

	if serverPort, exists := os.LookupEnv("CART_SERVER_PORT"); exists {
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

	return cfg
}
