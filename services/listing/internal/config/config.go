package config

import "os"

type Config struct {
	DatabaseURL string
	GRPCPort    string
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

	return Config{
		DatabaseURL: databaseURL,
		GRPCPort:    grpcPort,
	}
}