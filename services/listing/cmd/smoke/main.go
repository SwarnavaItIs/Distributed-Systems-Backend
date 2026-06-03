package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/swarnava/dmb/services/listing/internal/config"
	"github.com/swarnava/dmb/services/listing/internal/db"
	"github.com/swarnava/dmb/services/listing/internal/model"
	"github.com/swarnava/dmb/services/listing/internal/repository"
)

func main() {
	fmt.Println("Running Listing Repository smoke test...")

	cfg := config.Load()

	pool, err := db.NewPostgresPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	repo := repository.NewListingRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	listing := &model.Listing{
		SellerID:    "11111111-1111-1111-1111-111111111111",
		Title:       "iPhone 13",
		Description: "Used iPhone 13 in good condition",
		CategoryID:  1,
		PriceCents:  4500000,
		Status:      "ACTIVE",
	}

	err = repo.CreateListing(ctx, listing)
	if err != nil {
		log.Fatalf("create listing failed: %v", err)
	}

	fmt.Println("Created listing:")
	fmt.Println("ID:", listing.ID)
	fmt.Println("Title:", listing.Title)
	fmt.Println("Price cents:", listing.PriceCents)

	foundListing, err := repo.GetListing(ctx, listing.ID)
	if err != nil {
		log.Fatalf("get listing failed: %v", err)
	}

	fmt.Println("Fetched listing:")
	fmt.Println("ID:", foundListing.ID)
	fmt.Println("Seller ID:", foundListing.SellerID)
	fmt.Println("Title:", foundListing.Title)
	fmt.Println("Description:", foundListing.Description)
	fmt.Println("Category ID:", foundListing.CategoryID)
	fmt.Println("Price cents:", foundListing.PriceCents)
	fmt.Println("Status:", foundListing.Status)
}