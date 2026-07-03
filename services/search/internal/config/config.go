package config

import "os"

type Config struct {
	HTTPPort string
}

func Load() Config {
	httpPort := os.Getenv("SEARCH_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8081"
	}

	return Config{
		HTTPPort: httpPort,
	}
}