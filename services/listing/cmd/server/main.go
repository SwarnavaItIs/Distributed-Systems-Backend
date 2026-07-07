package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	listingv1 "github.com/swarnava/dmb/gen/go/listing/v1"
	"github.com/swarnava/dmb/services/listing/internal/config"
	"github.com/swarnava/dmb/services/listing/internal/db"
	"github.com/swarnava/dmb/services/listing/internal/events"
	"github.com/swarnava/dmb/services/listing/internal/repository"
	listingservice "github.com/swarnava/dmb/services/listing/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	fmt.Println("DMB Listing Service starting...")

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

	eventPublisher := events.NewPublisher(
		cfg.RedisAddr,
		cfg.ListingCreatedEventChannel,
	)

	if err := eventPublisher.Ping(startupCtx); err != nil {
		startupCancel()
		log.Fatalf(
			"redis connection failed for listing events: %v",
			err,
		)
	}

	startupCancel()
	defer eventPublisher.Close()

	fmt.Println("Redis connected successfully for listing events")
	fmt.Println(
		"Publishing listing events to channel:",
		cfg.ListingCreatedEventChannel,
	)

	listener, err := net.Listen(
		"tcp",
		":"+cfg.GRPCPort,
	)
	if err != nil {
		log.Fatalf(
			"failed to listen on port %s: %v",
			cfg.GRPCPort,
			err,
		)
	}

	repo := repository.NewListingRepository(pool)

	listingSvc := listingservice.NewListingService(
		repo,
		eventPublisher,
	)

	grpcServer := grpc.NewServer()

	listingv1.RegisterListingServiceServer(
		grpcServer,
		listingSvc,
	)

	reflection.Register(grpcServer)

	serverErr := make(chan error, 1)

	go func() {
		fmt.Println(
			"Listing gRPC server listening on port:",
			cfg.GRPCPort,
		)

		serverErr <- grpcServer.Serve(listener)
	}()

	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	select {
	case <-signalCtx.Done():
		fmt.Println("Listing Service shutdown signal received")

	case err := <-serverErr:
		if err != nil {
			log.Printf("Listing gRPC server stopped: %v", err)
		}
	}

	gracefulStopFinished := make(chan struct{})

	go func() {
		grpcServer.GracefulStop()
		close(gracefulStopFinished)
	}()

	select {
	case <-gracefulStopFinished:
		fmt.Println("Listing gRPC requests completed")

	case <-time.After(10 * time.Second):
		fmt.Println("gRPC graceful shutdown timed out")
		grpcServer.Stop()
	}

	fmt.Println("Listing Service stopped gracefully")
}
