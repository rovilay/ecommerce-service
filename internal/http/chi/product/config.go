package product

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort uint16
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
}

func LoadConfig() Config {
	cfg := Config{
		ServerPort: 3000,
	}

	if serverPort, exists := os.LookupEnv("PRODUCT_SERVER_PORT"); exists {
		if port, err := strconv.ParseUint(serverPort, 10, 16); err == nil {
			cfg.ServerPort = uint16(port)
		}
	}

	if url, exists := os.LookupEnv("DB_HOST"); exists {
		cfg.DBHost = url
	}

	if url, exists := os.LookupEnv("DB_PORT"); exists {
		cfg.DBPort = url
	}

	if url, exists := os.LookupEnv("DB_USER"); exists {
		cfg.DBUser = url
	}

	if url, exists := os.LookupEnv("DB_PASSWORD"); exists {
		cfg.DBPassword = url
	}

	return cfg
}
