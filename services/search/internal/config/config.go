package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort       string
	DatabaseURL    string
	RedisAddr      string
	SearchCacheTTL time.Duration
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

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	cacheTTLSeconds := 5
	if value := os.Getenv("SEARCH_CACHE_TTL_SECONDS"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil && parsed > 0 {
			cacheTTLSeconds = parsed
		}
	}

	return Config{
		HTTPPort:       httpPort,
		DatabaseURL:    databaseURL,
		RedisAddr:      redisAddr,
		SearchCacheTTL: time.Duration(cacheTTLSeconds) * time.Second,
	}
}