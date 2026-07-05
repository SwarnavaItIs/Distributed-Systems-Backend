package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/swarnava/dmb/services/gateway/internal/config"
	"github.com/swarnava/dmb/services/gateway/internal/handler"
	"github.com/swarnava/dmb/services/gateway/internal/middleware"
)

func main() {
	fmt.Println("DMB API Gateway starting...")

	cfg := config.Load()

	gatewayHandler, err := handler.NewGatewayHandler(cfg.SearchServiceURL)
	if err != nil {
		log.Fatalf("failed to create gateway handler: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", gatewayHandler.HealthHandler)

	mux.HandleFunc(
		"/api/search",
		middleware.JWTAuth(cfg.JWTSecret, gatewayHandler.SearchProxyHandler),
	)

	addr := ":" + cfg.HTTPPort

	fmt.Println("API Gateway listening on port:", cfg.HTTPPort)
	fmt.Println("Proxying search requests to:", cfg.SearchServiceURL)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("failed to start api gateway: %v", err)
	}
}