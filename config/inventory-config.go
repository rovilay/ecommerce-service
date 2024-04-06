package config

import (
	"os"
	"strconv"
)

type InventoryConfig struct {
	ServerPort uint16
	DBURL      string
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

	if url, exists := os.LookupEnv("DB_URL"); exists {
		cfg.DBURL = url
	}

	return cfg
}
