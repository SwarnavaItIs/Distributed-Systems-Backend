package model

import "time"

type SearchFilters struct {
	CategoryID int64 `json:"category_id,omitempty"`
	MinPrice   int64 `json:"min_price,omitempty"`
	MaxPrice   int64 `json:"max_price,omitempty"`
	Limit      int32 `json:"limit,omitempty"`
}

type SearchListing struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	CategoryID int64     `json:"category_id"`
	PriceCents int64     `json:"price_cents"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}
