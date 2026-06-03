package model

import "time"

type Listing struct {
	ID          string
	SellerID    string
	Title       string
	Description string
	CategoryID  int64
	PriceCents  int64
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}