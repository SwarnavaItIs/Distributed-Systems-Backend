package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	notificationHandler := handler.NewNotificationHandler(
		manager,
	)

	startupCtx, startupCancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	subscriber := notificationredis.NewSubscriber(
		cfg.RedisAddr,
		cfg.RedisChannel,
		manager,
	)

	if err := subscriber.Ping(startupCtx); err != nil {
		startupCancel()
		log.Fatalf(
			"redis connection failed for notification service: %v",
			err,
		)
	}

	startupCancel()
	defer subscriber.Close()

	fmt.Println(
		"Redis connected successfully for notification service",
	)

	appCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	subscriberErr := make(chan error, 1)

	go func() {
		subscriberErr <- subscriber.Start(appCtx)
	}()

	mux := http.NewServeMux()

	mux.HandleFunc(
		"/health",
		notificationHandler.HealthHandler,
	)

	mux.HandleFunc(
		"/ws",
		notificationHandler.WebSocketHandler,
	)

	mux.HandleFunc(
		"/broadcast",
		notificationHandler.BroadcastHandler,
	)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)

	go func() {
		fmt.Println(
			"Notification HTTP server listening on port:",
			cfg.HTTPPort,
		)

		fmt.Println(
			"WebSocket endpoint: ws://localhost:" +
				cfg.HTTPPort +
				"/ws",
		)

		fmt.Println(
			"Listening to Redis channel:",
			cfg.RedisChannel,
		)

		serverErr <- server.ListenAndServe()
	}()

	select {
	case <-appCtx.Done():
		fmt.Println(
			"Notification Service shutdown signal received",
		)

	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf(
				"Notification HTTP server stopped unexpectedly: %v",
				err,
			)
		}

	case err := <-subscriberErr:
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf(
				"Redis subscriber stopped unexpectedly: %v",
				err,
			)
		}
	}

	stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf(
			"Notification HTTP shutdown failed: %v",
			err,
		)

		if closeErr := server.Close(); closeErr != nil {
			log.Printf(
				"forced Notification HTTP close failed: %v",
				closeErr,
			)
		}
	}

	manager.CloseAll()

	fmt.Println("Notification Service stopped gracefully")
}
