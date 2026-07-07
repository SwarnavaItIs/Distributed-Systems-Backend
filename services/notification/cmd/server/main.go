package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/swarnava/dmb/services/notification/internal/config"
	"github.com/swarnava/dmb/services/notification/internal/handler"
	notificationredis "github.com/swarnava/dmb/services/notification/internal/redis"
	notificationws "github.com/swarnava/dmb/services/notification/internal/ws"
)

func main() {
	fmt.Println("DMB Notification Service starting...")

	cfg := config.Load()

	manager := notificationws.NewManager()
	notificationHandler := handler.NewNotificationHandler(manager)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	subscriber := notificationredis.NewSubscriber(
		cfg.RedisAddr,
		cfg.RedisChannel,
		manager,
	)

	if err := subscriber.Ping(ctx); err != nil {
		log.Fatalf("redis connection failed for notification service: %v", err)
	}
	defer subscriber.Close()

	fmt.Println("Redis connected successfully for notification service")

	go func() {
		if err := subscriber.Start(context.Background()); err != nil {
			log.Println("redis subscriber stopped:", err)
		}
	}()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", notificationHandler.HealthHandler)
	mux.HandleFunc("/ws", notificationHandler.WebSocketHandler)
	mux.HandleFunc("/broadcast", notificationHandler.BroadcastHandler)

	addr := ":" + cfg.HTTPPort

	fmt.Println("Notification HTTP server listening on port:", cfg.HTTPPort)
	fmt.Println("WebSocket endpoint: ws://localhost:" + cfg.HTTPPort + "/ws")
	fmt.Println("Listening to Redis channel:", cfg.RedisChannel)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("failed to start notification service: %v", err)
	}
}
