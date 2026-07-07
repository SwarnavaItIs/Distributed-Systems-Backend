package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

const testSellerID = "11111111-1111-1111-1111-111111111111"

type createListingRequest struct {
	SellerID    string `json:"seller_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CategoryID  int64  `json:"category_id"`
	PriceCents  int64  `json:"price_cents"`
}

type createListingResponse struct {
	Listing struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"listing"`
}

type websocketEvent struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	ListingID string `json:"listing_id"`
	Title     string `json:"title"`
}

func main() {
	gatewayURL := getEnv("GATEWAY_URL", "http://localhost:8080")
	websocketURL := getEnv("NOTIFICATION_WS_URL", "ws://localhost:8082/ws")
	jwtSecret := getEnv("JWT_SECRET", "dev_secret_change_me")

	token, err := generateToken(jwtSecret)
	if err != nil {
		log.Fatalf("failed to generate JWT: %v", err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(websocketURL, nil)
	if err != nil {
		log.Fatalf("failed to connect to Notification Service: %v", err)
	}
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		log.Fatalf("failed to set welcome message deadline: %v", err)
	}

	_, welcomeData, err := conn.ReadMessage()
	if err != nil {
		log.Fatalf("failed to read WebSocket welcome message: %v", err)
	}

	var welcome websocketEvent
	if err := json.Unmarshal(welcomeData, &welcome); err != nil {
		log.Fatalf("failed to decode welcome message: %v", err)
	}

	if welcome.Type != "connection.established" {
		log.Fatalf("unexpected welcome event type: %s", welcome.Type)
	}

	fmt.Println("Connected to Notification Service")

	title := fmt.Sprintf("DMB E2E Listing %d", time.Now().UnixNano())

	listingRequest := createListingRequest{
		SellerID:    testSellerID,
		Title:       title,
		Description: "Listing created by the DMB real-time end-to-end test",
		CategoryID:  8,
		PriceCents:  199900,
	}

	createdListing, err := createListing(gatewayURL, token, listingRequest)
	if err != nil {
		log.Fatalf("failed to create listing: %v", err)
	}

	fmt.Println("Created listing:", createdListing.Listing.ID)

	if err := conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		log.Fatalf("failed to set event deadline: %v", err)
	}

	for {
		_, eventData, err := conn.ReadMessage()
		if err != nil {
			log.Fatalf("failed while waiting for listing.created event: %v", err)
		}

		var event websocketEvent
		if err := json.Unmarshal(eventData, &event); err != nil {
			fmt.Println("Ignoring invalid WebSocket message:", string(eventData))
			continue
		}

		if event.Type != "listing.created" {
			fmt.Println("Ignoring unrelated event:", event.Type)
			continue
		}

		if event.ListingID != createdListing.Listing.ID {
			fmt.Printf(
				"Ignoring listing.created event for another listing: %s\n",
				event.ListingID,
			)
			continue
		}

		if event.Title != createdListing.Listing.Title {
			log.Fatalf(
				"title mismatch: created=%q event=%q",
				createdListing.Listing.Title,
				event.Title,
			)
		}

		fmt.Println("Received listing.created event:", event.ListingID)
		fmt.Println("PASS: real-time listing notification flow works")
		return
	}
}

func createListing(
	gatewayURL string,
	token string,
	payload createListingRequest,
) (*createListingResponse, error) {
	requestData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode listing request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		gatewayURL+"/api/listings",
		bytes.NewReader(requestData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gateway request failed: %w", err)
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read gateway response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf(
			"gateway returned status %d: %s",
			resp.StatusCode,
			string(responseData),
		)
	}

	var response createListingResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to decode gateway response: %w", err)
	}

	if response.Listing.ID == "" {
		return nil, fmt.Errorf("gateway response did not contain a listing id")
	}

	return &response, nil
}

func generateToken(secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  testSellerID,
		"role": "buyer",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func getEnv(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	return value
}
