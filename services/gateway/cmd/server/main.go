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

	"github.com/swarnava/dmb/services/gateway/internal/client"
	"github.com/swarnava/dmb/services/gateway/internal/config"
	"github.com/swarnava/dmb/services/gateway/internal/handler"
	"github.com/swarnava/dmb/services/gateway/internal/middleware"
)

func main() {
	fmt.Println("DMB API Gateway starting...")

	cfg := config.Load()

	gatewayHandler, err := handler.NewGatewayHandler(
		cfg.SearchServiceURL,
	)
	if err != nil {
		log.Fatalf("failed to create gateway handler: %v", err)
	}

	startupCtx, startupCancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	listingClient, listingConn, err := client.NewListingClient(
		startupCtx,
		cfg.ListingServiceAddr,
	)
	if err != nil {
		startupCancel()
		log.Fatalf("failed to create listing client: %v", err)
	}

	defer listingConn.Close()

	fmt.Println(
		"Connected to Listing Service:",
		cfg.ListingServiceAddr,
	)

	listingHandler := handler.NewListingHandler(listingClient)

	rateLimiter := middleware.NewRateLimiter(
		cfg.RedisAddr,
		cfg.RateLimitMax,
		cfg.RateLimitWindow,
	)

	if err := rateLimiter.Ping(startupCtx); err != nil {
		startupCancel()
		log.Fatalf(
			"redis connection failed for rate limiter: %v",
			err,
		)
	}

	startupCancel()
	defer rateLimiter.Close()

	fmt.Println("Redis connected successfully for rate limiter")

	mux := http.NewServeMux()

	mux.HandleFunc("/health", gatewayHandler.HealthHandler)

	mux.HandleFunc(
		"/api/search",
		middleware.JWTAuth(
			cfg.JWTSecret,
			rateLimiter.Limit(
				gatewayHandler.SearchProxyHandler,
			),
		),
	)

	mux.HandleFunc(
		"/api/listings",
		middleware.JWTAuth(
			cfg.JWTSecret,
			rateLimiter.Limit(
				listingHandler.CreateListingHandler,
			),
		),
	)

	mux.HandleFunc(
		"/api/listings/",
		middleware.JWTAuth(
			cfg.JWTSecret,
			rateLimiter.Limit(
				listingHandler.GetListingHandler,
			),
		),
	)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)

	go func() {
		fmt.Println(
			"API Gateway listening on port:",
			cfg.HTTPPort,
		)
		fmt.Println(
			"Proxying search requests to:",
			cfg.SearchServiceURL,
		)
		fmt.Println(
			"Routing listing requests to:",
			cfg.ListingServiceAddr,
		)
		fmt.Println(
			"Rate limit:",
			cfg.RateLimitMax,
			"requests per",
			cfg.RateLimitWindow,
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
		fmt.Println("API Gateway shutdown signal received")

	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Gateway stopped unexpectedly: %v", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful Gateway shutdown failed: %v", err)

		if closeErr := server.Close(); closeErr != nil {
			log.Printf("forced Gateway close failed: %v", closeErr)
		}
	}

	fmt.Println("API Gateway stopped gracefully")
}
