package product

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort uint16
	DBURL      string
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

	if url, exists := os.LookupEnv("PRODUCT_DB_URL"); exists {
		cfg.DBURL = url
	} else if url, exists = os.LookupEnv("DB_URL"); exists {
		cfg.DBURL = url
	}

	return cfg
}
