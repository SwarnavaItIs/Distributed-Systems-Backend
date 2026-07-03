package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type SearchFilters struct {
	CategoryID int64 `json:"category_id,omitempty"`
	MinPrice   int64 `json:"min_price,omitempty"`
	MaxPrice   int64 `json:"max_price,omitempty"`
}

type SearchListing struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	CategoryID  int64  `json:"category_id"`
	PriceCents  int64  `json:"price_cents"`
	Status      string `json:"status"`
}

type SearchResponse struct {
	Filters SearchFilters   `json:"filters"`
	Results []SearchListing `json:"results"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "search",
	})
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
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

	response := SearchResponse{
		Filters: SearchFilters{
			CategoryID: categoryID,
			MinPrice:   minPrice,
			MaxPrice:   maxPrice,
		},
		Results: []SearchListing{},
	}

	writeJSON(w, http.StatusOK, response)
}

func parseOptionalInt64(value string) (int64, error) {
	if value == "" {
		return 0, nil
	}

	return strconv.ParseInt(value, 10, 64)
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