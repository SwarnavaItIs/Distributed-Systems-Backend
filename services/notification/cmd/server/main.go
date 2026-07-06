package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/swarnava/dmb/services/notification/internal/config"
	"github.com/swarnava/dmb/services/notification/internal/handler"
)

func main() {
	fmt.Println("DMB Notification Service starting...")

	cfg := config.Load()

	notificationHandler := handler.NewNotificationHandler()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", notificationHandler.HealthHandler)
	mux.HandleFunc("/ws", notificationHandler.WebSocketHandler)

	addr := ":" + cfg.HTTPPort

	fmt.Println("Notification HTTP server listening on port:", cfg.HTTPPort)
	fmt.Println("WebSocket endpoint: ws://localhost:" + cfg.HTTPPort + "/ws")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("failed to start notification service: %v", err)
	}
}