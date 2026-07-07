package service

import (
	"context"
	"errors"
	"fmt"

	listingv1 "github.com/swarnava/dmb/gen/go/listing/v1"
	"github.com/swarnava/dmb/services/listing/internal/model"
	"github.com/swarnava/dmb/services/listing/internal/repository"
	"github.com/swarnava/dmb/services/listing/internal/validator"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ListingCreatedPublisher interface {
	PublishListingCreated(ctx context.Context, listing *model.Listing) error
}

type ListingService struct {
	listingv1.UnimplementedListingServiceServer
	repo      *repository.ListingRepository
	publisher ListingCreatedPublisher
}

func NewListingService(
	repo *repository.ListingRepository,
	publisher ListingCreatedPublisher,
) *ListingService {
	return &ListingService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *ListingService) CreateListing(
	ctx context.Context,
	req *listingv1.CreateListingRequest,
) (*listingv1.CreateListingResponse, error) {
	if err := validator.ValidateCreateListingRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	listing := &model.Listing{
		SellerID:    req.SellerId,
		Title:       req.Title,
		Description: req.Description,
		CategoryID:  req.CategoryId,
		PriceCents:  req.PriceCents,
		Status:      "ACTIVE",
	}

	if err := s.repo.CreateListing(ctx, listing); err != nil {
		return nil, status.Error(codes.Internal, "failed to create listing")
	}

	if s.publisher != nil {
		if err := s.publisher.PublishListingCreated(ctx, listing); err != nil {
			fmt.Println("failed to publish listing.created event:", err)
		}
	}

	return &listingv1.CreateListingResponse{
		Listing: modelToProto(listing),
	}, nil
}

func (s *ListingService) GetListing(
	ctx context.Context,
	req *listingv1.GetListingRequest,
) (*listingv1.GetListingResponse, error) {
	if err := validator.ValidateGetListingRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	listing, err := s.repo.GetListing(ctx, req.Id)
	if err != nil {
		if errors.Is(err, repository.ErrListingNotFound) {
			return nil, status.Error(codes.NotFound, "listing not found")
		}

		return nil, status.Error(codes.Internal, "failed to get listing")
	}

	return &listingv1.GetListingResponse{
		Listing: modelToProto(listing),
	}, nil
}

func modelToProto(listing *model.Listing) *listingv1.Listing {
	return &listingv1.Listing{
		Id:          listing.ID,
		SellerId:    listing.SellerID,
		Title:       listing.Title,
		Description: listing.Description,
		CategoryId:  listing.CategoryID,
		PriceCents:  listing.PriceCents,
		Status:      listing.Status,
		CreatedAt:   listing.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   listing.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
