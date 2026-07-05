package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rateLimiter := middleware.NewRateLimiter(
		cfg.RedisAddr,
		cfg.RateLimitMax,
		cfg.RateLimitWindow,
	)

	if err := rateLimiter.Ping(ctx); err != nil {
		log.Fatalf("redis connection failed for rate limiter: %v", err)
	}
	defer rateLimiter.Close()

	fmt.Println("Redis connected successfully for rate limiter")

	mux := http.NewServeMux()

	mux.HandleFunc("/health", gatewayHandler.HealthHandler)

	protectedSearchHandler := middleware.JWTAuth(
		cfg.JWTSecret,
		rateLimiter.Limit(gatewayHandler.SearchProxyHandler),
	)

	mux.HandleFunc("/api/search", protectedSearchHandler)

	addr := ":" + cfg.HTTPPort

	fmt.Println("API Gateway listening on port:", cfg.HTTPPort)
	fmt.Println("Proxying search requests to:", cfg.SearchServiceURL)
	fmt.Println("Rate limit:", cfg.RateLimitMax, "requests per", cfg.RateLimitWindow)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("failed to start api gateway: %v", err)
	}
}