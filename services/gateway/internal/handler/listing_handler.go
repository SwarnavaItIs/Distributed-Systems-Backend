package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	listingv1 "github.com/swarnava/dmb/gen/go/listing/v1"

	"google.golang.org/grpc/status"
)

type ListingHandler struct {
	client listingv1.ListingServiceClient
}

func NewListingHandler(client listingv1.ListingServiceClient) *ListingHandler {
	return &ListingHandler{
		client: client,
	}
}

type createListingHTTPRequest struct {
	SellerID    string `json:"seller_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CategoryID  int64  `json:"category_id"`
	PriceCents  int64  `json:"price_cents"`
}

func (h *ListingHandler) CreateListingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req createListingHTTPRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	resp, err := h.client.CreateListing(ctx, &listingv1.CreateListingRequest{
		SellerId:    req.SellerID,
		Title:       req.Title,
		Description: req.Description,
		CategoryId:  req.CategoryID,
		PriceCents:  req.PriceCents,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *ListingHandler) GetListingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/listings/")
	if id == "" || id == r.URL.Path {
		writeError(w, http.StatusBadRequest, "listing id is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	resp, err := h.client.GetListing(ctx, &listingv1.GetListingRequest{
		Id: id,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	switch st.Code().String() {
	case "InvalidArgument":
		writeError(w, http.StatusBadRequest, st.Message())
	case "NotFound":
		writeError(w, http.StatusNotFound, st.Message())
	default:
		writeError(w, http.StatusInternalServerError, st.Message())
	}
}