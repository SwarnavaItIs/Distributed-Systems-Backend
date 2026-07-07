package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/swarnava/dmb/services/listing/internal/model"
)

const ListingCreatedEventType = "listing.created"

type Publisher struct {
	client  *redis.Client
	channel string
}

type ListingCreatedEvent struct {
	Type       string `json:"type"`
	ListingID  string `json:"listing_id"`
	SellerID   string `json:"seller_id"`
	Title      string `json:"title"`
	CategoryID int64  `json:"category_id"`
	PriceCents int64  `json:"price_cents"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

func NewPublisher(redisAddr string, channel string) *Publisher {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &Publisher{
		client:  client,
		channel: channel,
	}
}

func (p *Publisher) Ping(ctx context.Context) error {
	return p.client.Ping(ctx).Err()
}

func (p *Publisher) Close() error {
	return p.client.Close()
}

func (p *Publisher) PublishListingCreated(
	ctx context.Context,
	listing *model.Listing,
) error {
	event := ListingCreatedEvent{
		Type:       ListingCreatedEventType,
		ListingID:  listing.ID,
		SellerID:   listing.SellerID,
		Title:      listing.Title,
		CategoryID: listing.CategoryID,
		PriceCents: listing.PriceCents,
		Status:     listing.Status,
		CreatedAt:  listing.CreatedAt.Format(time.RFC3339),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal listing created event: %w", err)
	}

	if err := p.client.Publish(ctx, p.channel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish listing created event: %w", err)
	}

	return nil
}