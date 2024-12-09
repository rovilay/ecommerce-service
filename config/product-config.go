package config

import (
	"os"
	"strconv"
)

type ProductConfig struct {
	ServerPort        uint16
	DBURL             string
	RABBITMQ_USER     string
	RABBITMQ_PASSWORD string
	RABBITMQ_PORT     uint16
	RABBITMQ_HOST     string
	RABBITMQ_URL      string
}

func LoadProductConfig() ProductConfig {
	cfg := ProductConfig{
		ServerPort: 3000,
	}

	if serverPort, exists := os.LookupEnv("PRODUCT_SERVER_PORT"); exists {
		if port, err := strconv.ParseUint(serverPort, 10, 16); err == nil {
			cfg.ServerPort = uint16(port)
		}
	}

	if rabbitMqDefaultUser, exists := os.LookupEnv("RABBITMQ_DEFAULT_USER"); exists {
		cfg.RABBITMQ_USER = rabbitMqDefaultUser
	}
	if rabbitMqDefaultPass, exists := os.LookupEnv("RABBITMQ_DEFAULT_PASS"); exists {
		cfg.RABBITMQ_PASSWORD = rabbitMqDefaultPass
	}
	if host, exists := os.LookupEnv("RABBITMQ_HOST"); exists {
		cfg.RABBITMQ_HOST = host
	}
	if rabbitmqUrl, exists := os.LookupEnv("RABBITMQ_URL"); exists {
		cfg.RABBITMQ_URL = rabbitmqUrl
	}
	if rabbitmqPort, exists := os.LookupEnv("RABBITMQ_PORT"); exists {
		if port, err := strconv.ParseUint(rabbitmqPort, 10, 16); err == nil {
			cfg.RABBITMQ_PORT = uint16(port)
		}
	}

	if url, exists := os.LookupEnv("DB_URL"); exists {
		cfg.DBURL = url
	}

	return cfg
}
