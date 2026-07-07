package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type GatewayHandler struct {
	searchProxy *httputil.ReverseProxy
}

func NewGatewayHandler(searchServiceURL string) (*GatewayHandler, error) {
	targetURL, err := url.Parse(searchServiceURL)
	if err != nil {
		return nil, fmt.Errorf("invalid search service url: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	originalDirector := proxy.Director

	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = "/search"
		req.Host = targetURL.Host
	}

	return &GatewayHandler{
		searchProxy: proxy,
	}, nil
}

func (h *GatewayHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "gateway",
	})
}

func (h *GatewayHandler) SearchProxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	h.searchProxy.ServeHTTP(w, r)
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
