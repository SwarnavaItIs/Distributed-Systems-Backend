package client

import (
	"context"
	"fmt"
	"time"

	listingv1 "github.com/swarnava/dmb/gen/go/listing/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func NewListingClient(
	ctx context.Context,
	listingServiceAddr string,
) (listingv1.ListingServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.DialContext(
		ctx,
		listingServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                15 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial listing service: %w", err)
	}

	client := listingv1.NewListingServiceClient(conn)

	return client, conn, nil
}