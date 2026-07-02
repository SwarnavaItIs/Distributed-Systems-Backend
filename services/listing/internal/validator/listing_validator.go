package validator

import (
	"errors"
	"strings"

	"github.com/google/uuid"

	listingv1 "github.com/swarnava/dmb/gen/go/listing/v1"
)

func ValidateCreateListingRequest(req *listingv1.CreateListingRequest) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}

	if strings.TrimSpace(req.SellerId) == "" {
		return errors.New("seller_id is required")
	}

	if _, err := uuid.Parse(req.SellerId); err != nil {
		return errors.New("seller_id must be a valid UUID")
	}

	if strings.TrimSpace(req.Title) == "" {
		return errors.New("title is required")
	}

	if len(req.Title) > 120 {
		return errors.New("title cannot be longer than 120 characters")
	}

	if len(req.Description) > 2000 {
		return errors.New("description cannot be longer than 2000 characters")
	}

	if req.CategoryId <= 0 {
		return errors.New("category_id must be greater than 0")
	}

	if req.PriceCents <= 0 {
		return errors.New("price_cents must be greater than 0")
	}

	return nil
}

func ValidateGetListingRequest(req *listingv1.GetListingRequest) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}

	if strings.TrimSpace(req.Id) == "" {
		return errors.New("id is required")
	}

	if _, err := uuid.Parse(req.Id); err != nil {
		return errors.New("id must be a valid UUID")
	}

	return nil
}
