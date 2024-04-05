package config

import (
	"log"
	"os"
	"strconv"
)

type InventoryConfig struct {
	ProductBaseUrl string
	ServerPort     uint16
	DBURL          string
}

func LoadInventoryConfig() InventoryConfig {
	cfg := InventoryConfig{
		ServerPort: 3000,
	}

	if serverPort, exists := os.LookupEnv("INVENTORY_SERVER_PORT"); exists {
		if port, err := strconv.ParseUint(serverPort, 10, 16); err == nil {
			cfg.ServerPort = uint16(port)
		}
	}

	if url, exists := os.LookupEnv("INVENTORY_DB_URL"); exists {
		cfg.DBURL = url
	} else if url, exists = os.LookupEnv("DB_URL"); exists {
		cfg.DBURL = url
	}

	url, exists := os.LookupEnv("PRODUCT_BASE_URL")

	if exists {
		cfg.ProductBaseUrl = url
	} else {
		log.Fatal("ProductBaseUrl required for this service")
	}

	return cfg
}
