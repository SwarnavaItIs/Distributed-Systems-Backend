package config

import "os"

type Config struct {
	HTTPPort         string
	SearchServiceURL string
	JWTSecret        string
}

func Load() Config {
	httpPort := os.Getenv("GATEWAY_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	searchServiceURL := os.Getenv("SEARCH_SERVICE_URL")
	if searchServiceURL == "" {
		searchServiceURL = "http://localhost:8081"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev_secret_change_me"
	}

	return Config{
		HTTPPort:         httpPort,
		SearchServiceURL: searchServiceURL,
		JWTSecret:        jwtSecret,
	}
}