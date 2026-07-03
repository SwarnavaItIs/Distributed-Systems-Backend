package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/swarnava/dmb/services/search/internal/model"
	"github.com/swarnava/dmb/services/search/internal/repository"
)

type SearchHandler struct {
	repo *repository.SearchRepository
}

func NewSearchHandler(repo *repository.SearchRepository) *SearchHandler {
	return &SearchHandler{
		repo: repo,
	}
}

type SearchResponse struct {
	Filters model.SearchFilters   `json:"filters"`
	Results []model.SearchListing `json:"results"`
	Count   int                   `json:"count"`
}

func (h *SearchHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "search",
	})
}

func (h *SearchHandler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query := r.URL.Query()

	categoryID, err := parseOptionalInt64(query.Get("category_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "category_id must be a valid integer")
		return
	}

	minPrice, err := parseOptionalInt64(query.Get("min_price"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "min_price must be a valid integer")
		return
	}

	maxPrice, err := parseOptionalInt64(query.Get("max_price"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "max_price must be a valid integer")
		return
	}

	limit, err := parseOptionalInt32(query.Get("limit"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "limit must be a valid integer")
		return
	}

	if minPrice > 0 && maxPrice > 0 && minPrice > maxPrice {
		writeError(w, http.StatusBadRequest, "min_price cannot be greater than max_price")
		return
	}

	filters := model.SearchFilters{
		CategoryID: categoryID,
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
		Limit:      limit,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	results, err := h.repo.SearchListings(ctx, filters)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to search listings")
		return
	}

	response := SearchResponse{
		Filters: filters,
		Results: results,
		Count:   len(results),
	}

	writeJSON(w, http.StatusOK, response)
}

func parseOptionalInt64(value string) (int64, error) {
	if value == "" {
		return 0, nil
	}

	return strconv.ParseInt(value, 10, 64)
}

func parseOptionalInt32(value string) (int32, error) {
	if value == "" {
		return 20, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(parsed), nil
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{
		"error": message,
	})
}