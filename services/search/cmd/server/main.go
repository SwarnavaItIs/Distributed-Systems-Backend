package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/swarnava/dmb/services/search/internal/config"
	"github.com/swarnava/dmb/services/search/internal/handler"
)

func main() {
	fmt.Println("DMB Search Service starting...")

	cfg := config.Load()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.HealthHandler)
	mux.HandleFunc("/search", handler.SearchHandler)

	addr := ":" + cfg.HTTPPort

	fmt.Println("Search HTTP server listening on port:", cfg.HTTPPort)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("failed to start search service: %v", err)
	}
}