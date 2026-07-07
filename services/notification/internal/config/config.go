package config

import "os"

type Config struct {
	HTTPPort     string
	RedisAddr    string
	RedisChannel string
}

func Load() Config {
	httpPort := os.Getenv("NOTIFICATION_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8082"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisChannel := os.Getenv("REDIS_LISTING_CREATED_CHANNEL")
	if redisChannel == "" {
		redisChannel = "listing.created"
	}

	return Config{
		HTTPPort:     httpPort,
		RedisAddr:    redisAddr,
		RedisChannel: redisChannel,
	}
}
