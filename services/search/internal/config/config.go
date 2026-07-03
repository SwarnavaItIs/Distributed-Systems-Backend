package config

import "os"

type Config struct {
	HTTPPort    string
	DatabaseURL string
}

func Load() Config {
	httpPort := os.Getenv("SEARCH_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8081"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://dmb_user:dmb_password@localhost:5432/dmb?sslmode=disable"
	}

	return Config{
		HTTPPort:    httpPort,
		DatabaseURL: databaseURL,
	}
}