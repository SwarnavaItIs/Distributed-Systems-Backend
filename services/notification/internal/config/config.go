package config

import "os"

type Config struct {
	HTTPPort string
}

func Load() Config {
	httpPort := os.Getenv("NOTIFICATION_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8082"
	}

	return Config{
		HTTPPort: httpPort,
	}
}