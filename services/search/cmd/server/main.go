package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	searchcache "github.com/swarnava/dmb/services/search/internal/cache"
	"github.com/swarnava/dmb/services/search/internal/config"
	"github.com/swarnava/dmb/services/search/internal/db"
	"github.com/swarnava/dmb/services/search/internal/handler"
	"github.com/swarnava/dmb/services/search/internal/repository"
)

func main() {
	fmt.Println("DMB Search Service starting...")

	cfg := config.Load()

	pool, err := db.NewPostgresPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	fmt.Println("Database connected successfully")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	searchCache := searchcache.NewSearchCache(cfg.RedisAddr, cfg.SearchCacheTTL)
	if err := searchCache.Ping(ctx); err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	defer searchCache.Close()

	fmt.Println("Redis connected successfully")

	searchRepo := repository.NewSearchRepository(pool)
	searchHandler := handler.NewSearchHandler(searchRepo, searchCache)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", searchHandler.HealthHandler)
	mux.HandleFunc("/search", searchHandler.SearchHandler)

	addr := ":" + cfg.HTTPPort

	fmt.Println("Search HTTP server listening on port:", cfg.HTTPPort)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("failed to start search service: %v", err)
	}
}