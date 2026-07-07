package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort           string
	SearchServiceURL   string
	ListingServiceAddr string
	JWTSecret          string
	RedisAddr          string
	RateLimitMax       int64
	RateLimitWindow    time.Duration
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

	listingServiceAddr := os.Getenv("LISTING_SERVICE_ADDR")
	if listingServiceAddr == "" {
		listingServiceAddr = "localhost:50051"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev_secret_change_me"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rateLimitMax := int64(100)
	if value := os.Getenv("RATE_LIMIT_MAX"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err == nil && parsed > 0 {
			rateLimitMax = parsed
		}
	}

	rateLimitWindowSeconds := 60
	if value := os.Getenv("RATE_LIMIT_WINDOW_SECONDS"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil && parsed > 0 {
			rateLimitWindowSeconds = parsed
		}
	}

	return Config{
		HTTPPort:           httpPort,
		SearchServiceURL:   searchServiceURL,
		ListingServiceAddr: listingServiceAddr,
		JWTSecret:          jwtSecret,
		RedisAddr:          redisAddr,
		RateLimitMax:       rateLimitMax,
		RateLimitWindow:    time.Duration(rateLimitWindowSeconds) * time.Second,
	}
}
