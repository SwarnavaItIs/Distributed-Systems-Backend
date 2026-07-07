package config

import "os"

type Config struct {
	DatabaseURL                string
	GRPCPort                   string
	RedisAddr                  string
	ListingCreatedEventChannel string
}

func Load() Config {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://dmb_user:dmb_password@localhost:5432/dmb?sslmode=disable"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	listingCreatedEventChannel := os.Getenv("REDIS_LISTING_CREATED_CHANNEL")
	if listingCreatedEventChannel == "" {
		listingCreatedEventChannel = "listing.created"
	}

	return Config{
		DatabaseURL:                databaseURL,
		GRPCPort:                   grpcPort,
		RedisAddr:                  redisAddr,
		ListingCreatedEventChannel: listingCreatedEventChannel,
	}
}