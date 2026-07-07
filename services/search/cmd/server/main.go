package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	startupCtx, startupCancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	searchCache := searchcache.NewSearchCache(
		cfg.RedisAddr,
		cfg.SearchCacheTTL,
	)

	if err := searchCache.Ping(startupCtx); err != nil {
		startupCancel()
		log.Fatalf("redis connection failed: %v", err)
	}

	startupCancel()
	defer searchCache.Close()

	fmt.Println("Redis connected successfully")

	searchRepo := repository.NewSearchRepository(pool)
	searchHandler := handler.NewSearchHandler(searchRepo, searchCache)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", searchHandler.HealthHandler)
	mux.HandleFunc("/search", searchHandler.SearchHandler)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)

	go func() {
		fmt.Println(
			"Search HTTP server listening on port:",
			cfg.HTTPPort,
		)

		serverErr <- server.ListenAndServe()
	}()

	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	select {
	case <-signalCtx.Done():
		fmt.Println("Search Service shutdown signal received")

	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("search HTTP server stopped unexpectedly: %v", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful HTTP shutdown failed: %v", err)

		if closeErr := server.Close(); closeErr != nil {
			log.Printf("forced HTTP close failed: %v", closeErr)
		}
	}

	fmt.Println("Search Service stopped gracefully")
}
