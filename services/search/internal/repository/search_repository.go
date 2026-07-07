package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/swarnava/dmb/services/search/internal/model"
)

type SearchRepository struct {
	db *pgxpool.Pool
}

func NewSearchRepository(db *pgxpool.Pool) *SearchRepository {
	return &SearchRepository{
		db: db,
	}
}

func (r *SearchRepository) SearchListings(
	ctx context.Context,
	filters model.SearchFilters,
) ([]model.SearchListing, error) {
	if filters.Limit <= 0 || filters.Limit > 100 {
		filters.Limit = 20
	}

	query := `
		SELECT
			id,
			title,
			category_id,
			price_cents,
			status,
			created_at
		FROM listings
		WHERE status = 'ACTIVE'
		AND ($1::BIGINT = 0 OR category_id = $1)
		AND ($2::BIGINT = 0 OR price_cents >= $2)
		AND ($3::BIGINT = 0 OR price_cents <= $3)
		ORDER BY price_cents ASC
		LIMIT $4;
	`

	rows, err := r.db.Query(
		ctx,
		query,
		filters.CategoryID,
		filters.MinPrice,
		filters.MaxPrice,
		filters.Limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search listings: %w", err)
	}
	defer rows.Close()

	listings := make([]model.SearchListing, 0)

	for rows.Next() {
		var listing model.SearchListing

		err := rows.Scan(
			&listing.ID,
			&listing.Title,
			&listing.CategoryID,
			&listing.PriceCents,
			&listing.Status,
			&listing.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan listing: %w", err)
		}

		listings = append(listings, listing)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}

	return listings, nil
}
