package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/swarnava/dmb/services/listing/internal/model"
)

var ErrListingNotFound = errors.New("listing not found")

type ListingRepository struct {
	db *pgxpool.Pool
}

func NewListingRepository(db *pgxpool.Pool) *ListingRepository {
	return &ListingRepository{
		db: db,
	}
}

func (r *ListingRepository) CreateListing(ctx context.Context, listing *model.Listing) error {
	query := `
		INSERT INTO listings (
			seller_id,
			title,
			description,
			category_id,
			price_cents,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at;
	`

	if listing.Status == "" {
		listing.Status = "ACTIVE"
	}

	err := r.db.QueryRow(
		ctx,
		query,
		listing.SellerID,
		listing.Title,
		listing.Description,
		listing.CategoryID,
		listing.PriceCents,
		listing.Status,
	).Scan(
		&listing.ID,
		&listing.CreatedAt,
		&listing.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create listing: %w", err)
	}

	return nil
}

func (r *ListingRepository) GetListing(ctx context.Context, id string) (*model.Listing, error) {
	query := `
		SELECT
			id,
			seller_id,
			title,
			COALESCE(description, '') AS description,
			category_id,
			price_cents,
			status,
			created_at,
			updated_at
		FROM listings
		WHERE id = $1;
	`

	var listing model.Listing

	err := r.db.QueryRow(ctx, query, id).Scan(
		&listing.ID,
		&listing.SellerID,
		&listing.Title,
		&listing.Description,
		&listing.CategoryID,
		&listing.PriceCents,
		&listing.Status,
		&listing.CreatedAt,
		&listing.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrListingNotFound
		}

		return nil, fmt.Errorf("failed to get listing: %w", err)
	}

	return &listing, nil
}