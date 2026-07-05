package config

import "os"

type Config struct {
	HTTPPort         string
	SearchServiceURL string
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

	return Config{
		HTTPPort:         httpPort,
		SearchServiceURL: searchServiceURL,
	}
}