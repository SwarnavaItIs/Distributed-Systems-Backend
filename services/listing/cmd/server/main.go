package main

import (
	"fmt"
	"log"
	"net"

	listingv1 "github.com/swarnava/dmb/gen/go/listing/v1"
	"github.com/swarnava/dmb/services/listing/internal/config"
	"github.com/swarnava/dmb/services/listing/internal/db"
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

	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

	repo := repository.NewListingRepository(pool)
	listingSvc := listingservice.NewListingService(repo)

	grpcServer := grpc.NewServer()

	listingv1.RegisterListingServiceServer(grpcServer, listingSvc)

	reflection.Register(grpcServer)

	fmt.Println("Listing gRPC server listening on port:", cfg.GRPCPort)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve gRPC server: %v", err)
	}
}
