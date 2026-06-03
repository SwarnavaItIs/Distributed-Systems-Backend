package main

import (
	"fmt"
	"log"

	"github.com/swarnava/dmb/services/listing/internal/config"
	"github.com/swarnava/dmb/services/listing/internal/db"
)

func main() {
	fmt.Println("Listing Service starting...")

	cfg := config.Load()

	pool, err := db.NewPostgresPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	fmt.Println("Database connected successfully")
	fmt.Println("Listing Service ready on port:", cfg.GRPCPort)
}